package containers

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/disk"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/settings"
	"github.com/threefoldtech/0-core/base/utils"
)

const (
	BackendBaseDir       = "/var/cache/containers"
	ContainerBaseRootDir = "/mnt/containers"
)

func (c *container) name() string {
	return fmt.Sprintf("%d", c.id)
}

func (c *container) flistConfigOverride(target string, cfg map[string]string) error {
	for name, content := range cfg {
		p := path.Join(target, utils.SafeNormalize(name))
		if err := os.MkdirAll(path.Dir(p), 0700); err != nil {
			return fmt.Errorf("failed to create director: %s", path.Dir(p))
		}
		if err := ioutil.WriteFile(p, []byte(content), 0600); err != nil {
			return err
		}
	}
	return nil
}

//mergeFList layers one (and only one) flist on top of the container root flist
//usually used for debugging
func (c *container) mergeFList(src string) error {
	arguments, err := c.Arguments()
	if err != nil {
		return err
	}
	namespace := fmt.Sprintf("containers/%s", c.name())
	return c.mgr.filesystem().MergeFList(namespace, c.root(), arguments.Root, src)
}

func (c *container) mountFList(storage, src string, target string, cfg map[string]string, hooks ...pm.RunnerHook) error {
	//check
	namespace := fmt.Sprintf("containers/%s/fs", c.name())

	err := c.mgr.filesystem().MountFList(namespace, storage, src, target, hooks...)
	if err != nil {
		return err
	}
	err = c.flistConfigOverride(target, cfg)
	if err != nil {
		return err
	}

	return nil
}

func (c *container) configFile() string {
	return path.Join(c.workingDir(), "config.json")
}

func (c *container) workingDir() string {
	return path.Join(BackendBaseDir, c.name())
}

func (c *container) fsDir() string {
	return path.Join(c.workingDir(), "fs")
}

func (c *container) root() string {
	return path.Join(ContainerBaseRootDir, c.name())
}

func (c *container) pipeIn() string {
	return path.Join(c.workingDir(), "in")
}

func (c *container) pipeOut() string {
	return path.Join(c.workingDir(), "out")
}

type SortableDisks []disk.PartitionStat

func (d SortableDisks) Len() int {
	return len(d)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (d SortableDisks) Less(i, j int) bool {
	return len(d[i].Mountpoint) > len(d[j].Mountpoint)
}

// Swap swaps the elements with indexes i and j.
func (d SortableDisks) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (c *container) getFSType(dir string) string {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}

	dir = strings.TrimRight(dir, "/") + "/"

	parts, err := disk.Partitions(true)
	if err != nil {
		return ""
	}

	sort.Sort(SortableDisks(parts))

	for _, part := range parts {
		mountpoint := part.Mountpoint
		if mountpoint != "/" {
			mountpoint += "/"
		}

		if strings.Index(dir, mountpoint) == 0 {
			return part.Fstype
		}
	}

	return ""
}

func (c *container) touch(p string) error {
	f, err := os.Create(p)

	if err != nil {
		return err
	}

	return f.Close()
}

func (c *container) sandbox(args *ContainerCreateArguments) error {
	//mount root flist.
	//prepare root folder.
	//make sure we remove the directory
	workDir := c.workingDir()
	os.RemoveAll(workDir)
	fstype := c.getFSType(BackendBaseDir)
	log.Debugf("Sandbox fileystem type: %s", fstype)

	if fstype == "btrfs" {
		//make sure we delete it if sub volume exists
		if utils.Exists(workDir) {
			c.mgr.api.System("btrfs", "subvolume", "delete", workDir)
		}
		c.mgr.api.System("btrfs", "subvolume", "create", workDir)
	}

	root := c.root()
	log.Debugf("Container root: %s", root)
	os.RemoveAll(root)

	onSBExit := &pm.ExitHook{
		Action: func(_ bool) {
			c.Terminate()
			c.cleanSandbox()
		},
	}

	storage := args.Storage
	if storage == "" {
		storage = settings.Settings.Globals.Get("storage", "zdb://hub.grid.tf:9900")
		args.Storage = storage
	}

	if err := c.mountFList(storage, args.Root, root, args.Config, onSBExit); err != nil {
		return fmt.Errorf("mount-root-flist(%s)", err)
	}

	os.MkdirAll(path.Join(root, "etc"), 0755)

	for src, dst := range args.Mount {
		target := path.Join(root, dst)

		//src can either be a location on HD, or another flist
		u, err := url.Parse(src)
		if err != nil {
			return fmt.Errorf("bad mount source '%s': %s", src, err)
		}

		if u.Scheme == "" {
			info, err := os.Stat(src)
			if err != nil {
				return err
			}
			if info.IsDir() {
				os.MkdirAll(target, 0755)
			} else {
				os.MkdirAll(path.Dir(target), 07555)
				if err := c.touch(target); err != nil {
					return err
				}
			}
			if err := syscall.Mount(src, target, "", syscall.MS_BIND, ""); err != nil {
				return fmt.Errorf("mount-bind(%s)", err)
			}
		} else {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			//assume a flist
			if err := c.mountFList(storage, src, target, nil); err != nil {
				return fmt.Errorf("mount-bind-flist(%s)", err)
			}
		}
	}

	coreXTarget := path.Join(root, coreXBinaryName)
	if err := c.touch(coreXTarget); err != nil {
		return err
	}

	coreXSrc, err := exec.LookPath(coreXBinaryName)
	if err != nil {
		return err
	}

	return syscall.Mount(coreXSrc, coreXTarget, "", syscall.MS_BIND, "")
}

func (c *container) cleanSandbox() {
	c.unMountAll()

	if c.getFSType(BackendBaseDir) == "btrfs" {
		c.mgr.api.System("btrfs", "subvolume", "delete", path.Join(BackendBaseDir, c.name()))
	} else {
		os.RemoveAll(path.Join(BackendBaseDir, c.name()))
	}

	os.RemoveAll(c.root())
}

func (c *container) unMountAll() error {
	mnts, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return err
	}
	root := c.root()
	var targets []string
	for _, line := range strings.Split(string(mnts), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		target := fields[1]
		if target == root || strings.HasPrefix(target, root+"/") {
			targets = append(targets, target)
		}
	}

	sort.Slice(targets, func(i, j int) bool {
		return strings.Count(targets[i], "/") > strings.Count(targets[j], "/")
	})

	for _, target := range targets {
		log.Debugf("unmounting '%s'", target)
		if err := syscall.Unmount(target, syscall.MNT_DETACH); err != nil {
			log.Errorf("failed to un-mount '%s': %s", target, err)
		}
	}

	return nil
}
