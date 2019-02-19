package zfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"syscall"

	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/settings"
	"github.com/threefoldtech/0-core/base/stream"
	"github.com/threefoldtech/0-core/base/utils"
)

const (
	CacheBaseDir    = "/var/cache"
	CacheZeroFSDir  = CacheBaseDir + "/zerofs"
	CacheFListDir   = CacheBaseDir + "/flist"
	LocalRouterFile = CacheBaseDir + "/router.yaml"
)

func (m *Manager) mount(ctx pm.Context) (interface{}, error) {
	var args struct {
		Namespace string `json:"namespace"`
		Storage   string `json:"storage"`
		Source    string `json:"source"`
		Target    string `json:"target"`
	}

	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	return nil, m.MountFList(args.Namespace, args.Storage, args.Source, args.Target)
}

func getNSID(ns string) string {
	return fmt.Sprintf("zfs:%s", ns)
}

func (m *Manager) MountFList(namespace, storage, src string, target string, hooks ...pm.RunnerHook) error {
	//check
	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	meta, err := getMeta(src)

	if err != nil {
		return err
	}

	backend := path.Join(CacheBaseDir, namespace, meta.Hash)
	os.RemoveAll(backend)
	os.MkdirAll(backend, 0755)

	cache := settings.Settings.Globals.Get("cache", CacheZeroFSDir)
	g8ufs := []string{
		"--reset",
		"--backend", backend,
		"--cache", cache,
		"--log", path.Join(backend, "fs.log"),
		"--meta", meta.Base,
		"--storage-url", storage,
	}

	//local router files
	if utils.Exists(LocalRouterFile) {
		g8ufs = append(g8ufs,
			"--local-router", LocalRouterFile,
		)
	}

	g8ufs = append(g8ufs, target)
	cmd := &pm.Command{
		ID:      path.Join(getNSID(namespace), target),
		Command: pm.CommandSystem,
		Arguments: pm.MustArguments(pm.SystemCommandArguments{
			Name: "g8ufs",
			Args: g8ufs,
		}),
	}

	var j pm.Job
	var o sync.Once
	var wg sync.WaitGroup
	wg.Add(1)

	hooks = append(hooks, &pm.MatchHook{
		Match: "mount starts",
		Action: func(_ *stream.Message) {
			o.Do(wg.Done)
		},
	}, &pm.ExitHook{
		Action: func(s bool) {
			o.Do(func() {
				if !s {
					result := j.Wait()
					err = fmt.Errorf("abnormal exit of filesystem mount at '%s': %s", target, result.Streams)
				}
				wg.Done()
			})
		},
	})

	j, err = m.api.Run(cmd, hooks...)
	if err != nil {
		return err
	}

	//wait for either of the hooks (ready or exit)
	wg.Wait()
	return err
}

// MergeFList layers the given  flist on top of the mounted flist(namespace, target, base). The fs must be running
// before your do the merge
// To prevent abuse, the MergeFList allows layering only ONE flist on top of the running fs. by overriding the
// fs `layered` file.
// the namespace, targe, and base are needed to identify the g8ufs process, the flist is the one to layer, where base
// is the original flist used on the call to MountFlist
func (m *Manager) MergeFList(namespace, target, base, flist string) error {
	id := path.Join(getNSID(namespace), target)
	job, ok := m.api.JobOf(id)
	if !ok {
		return fmt.Errorf("no filesystem running for the provided namespace and target (%s/%s)", namespace, target)
	}

	meta, err := getMeta(flist)
	if err != nil {
		return err
	}

	//append db to the backend/.layered file
	f, err := getFList(base)
	if err != nil {
		return err
	}
	hash, err := f.Hash()
	if err != nil {
		return err
	}

	baseBackend := path.Join(CacheBaseDir, namespace, hash)
	if err := ioutil.WriteFile(path.Join(baseBackend, ".layered"), []byte(meta.Base), 0644); err != nil {
		return err
	}

	return job.Signal(syscall.SIGHUP) //signal fs reload
}

func (m *Manager) GetCacheZeroFSDir() string {
	return CacheZeroFSDir
}
