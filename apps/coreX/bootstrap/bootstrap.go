package bootstrap

import (
	"fmt"
	"os"
	"path"
	"syscall"

	logging "github.com/op/go-logging"
	"github.com/shirou/gopsutil/process"
	"github.com/threefoldtech/0-core/apps/coreX/options"
	"github.com/threefoldtech/0-core/base/mgr"
	"github.com/threefoldtech/0-core/base/settings"
	"github.com/threefoldtech/0-core/base/utils"
)

var (
	log = logging.MustGetLogger("bootstrap")
)

type DeviceType uint32

const (
	CharDevice  DeviceType = syscall.S_IFCHR
	BlockDevice DeviceType = syscall.S_IFBLK
)

type Device struct {
	Name  string
	Type  DeviceType
	Mode  os.FileMode
	Major int
	Minor int
}

func (d *Device) mk(in string) error {
	return syscall.Mknod(path.Join(in, d.Name),
		uint32(d.Type)|uint32(d.Mode),
		d.Major<<8|d.Minor,
	)
}

type Bootstrap struct {
}

func NewBootstrap() *Bootstrap {
	return &Bootstrap{}
}

func (o *Bootstrap) populateMinimumDev() error {
	devices := []Device{
		{"console", CharDevice, 0600, 136, 2},
		{"full", CharDevice, 0666, 1, 7},
		{"null", CharDevice, 0666, 1, 3},
		{"random", CharDevice, 0666, 1, 8},
		{"tty", CharDevice, 0666, 5, 0},
		{"urandom", CharDevice, 0666, 1, 9},
		{"zero", CharDevice, 0666, 1, 5},
	}

	previousUmask := syscall.Umask(0000)

	for _, dev := range devices {
		if err := dev.mk("/dev"); err != nil {
			return fmt.Errorf("failed to create device %v: %s", dev, err)
		}
	}

	for oldname, newname := range map[string]string{
		"/proc/self/fd/0": "/dev/stdin",
		"/proc/self/fd/1": "/dev/stdout",
		"/proc/self/fd/2": "/dev/stderr",
		"/proc/self/fd":   "/dev/fd",
		"/dev/pts/ptmx":   "/dev/ptmx",
		"/proc/kcore":     "/dev/core",
	} {
		if err := os.Symlink(oldname, newname); err != nil {
			return fmt.Errorf("failed to create symlink %s->%s: %s", newname, oldname, err)
		}
	}

	os.MkdirAll("/dev/mqueue", 0777)
	if err := syscall.Mount("mqueue", "/dev/mqueue", "mqueue", 0, ""); err != nil {
		return fmt.Errorf("failed to mount mqueue: %s", err)
	}

	os.MkdirAll("/dev/shm", 0777)
	if err := syscall.Mount("shm", "/dev/shm", "tmpfs",
		syscall.MS_NOSUID|syscall.MS_NODEV|syscall.MS_NOEXEC|syscall.MS_RELATIME,
		"size=1G"); err != nil {
		return fmt.Errorf("failed to mount shm: %s", err)
	}

	syscall.Umask(previousUmask)

	return nil
}

func (o *Bootstrap) Start() error {
	log.Debugf("setting up mounts")
	os.MkdirAll("/etc", 0755)
	os.MkdirAll("/var/run", 0755)

	os.MkdirAll("/sys", 0755)
	if err := syscall.Mount("none", "/sys", "sysfs",
		syscall.MS_NOSUID|syscall.MS_RELATIME|syscall.MS_NODEV|syscall.MS_NOEXEC|syscall.MS_RDONLY,
		""); err != nil {
		return err
	}

	os.MkdirAll("/proc", 0755)
	procflags := uintptr(syscall.MS_NOSUID | syscall.MS_RELATIME | syscall.MS_NODEV | syscall.MS_NOEXEC)
	if options.Options.Unprivileged() {
		procflags |= syscall.MS_RDONLY
	}

	if err := syscall.Mount("none", "/proc", "proc", procflags, ""); err != nil {
		return err
	}

	os.MkdirAll("/dev", 0755)
	if options.Options.Unprivileged() {
		if err := syscall.Mount("none", "/dev", "tmpfs", syscall.MS_NOSUID, "mode=755"); err != nil {
			return fmt.Errorf("failed to mount dev in unprivileged: %s", err)
		}
		if err := o.populateMinimumDev(); err != nil {
			return err
		}
	} else {
		if err := syscall.Mount("none", "/dev", "devtmpfs", syscall.MS_NOSUID|syscall.MS_RELATIME, "mode=755"); err != nil {
			return err
		}
	}

	os.MkdirAll("/dev/pts", 0755)
	if err := syscall.Mount("none", "/dev/pts", "devpts", 0, ""); err != nil {
		return err
	}

	return nil
}

func (b *Bootstrap) startup() error {
	var included settings.IncludedSettings
	if err := utils.LoadTomlFile("/.startup.toml", &included); err != nil {
		return err
	}

	tree, errs := included.GetStartupTree()
	if errs != nil {
		return fmt.Errorf("failed to build startup tree: %v", errs)
	}

	mgr.RunSlice(tree.Slice(settings.AfterInit.Weight(), settings.ToTheEnd.Weight()))

	return nil
}

//Bootstrap registers extensions and startup system services.
func (b *Bootstrap) Bootstrap(hostname string) error {
	log.Debugf("setting up hostname")
	if err := updateHostname(hostname); err != nil {
		return err
	}

	log.Debugf("linkin mtab")
	if err := linkMtab(); err != nil {
		return err
	}

	if options.Options.Unprivileged() {
		mgr.SetUnprivileged()
		if err := b.revokePrivileges(); err != nil {
			return err
		}
	}

	log.Debugf("startup services")

	if err := b.plugins(); err != nil {
		log.Errorf("failed to load plugins: %s", err)
	}

	if err := b.startup(); err != nil {
		log.Errorf("failed to startup container services: %s", err)
	}

	return nil
}

func (b *Bootstrap) UnBootstrap() {
	//clean up behind (kill all processes)
	pids, _ := process.Pids()
	//kill all children.
	for _, pid := range pids {
		if pid == 1 {
			continue
		}
		log.Infof("stopping PID: %d", pid)
		ps, err := process.NewProcess(pid)

		if err != nil {
			log.Errorf("failed to kill pid (%d): %s", pid, err)
			continue
		}

		if err := ps.Kill(); err != nil {
			log.Errorf("failed to kill pid (%d): %s", pid, err)
			continue
		}
	}

	for _, mnt := range []string{"/dev/pts", "/dev", "proc"} {
		log.Infof("Unmounting: %s", mnt)
		if err := syscall.Unmount(mnt, syscall.MNT_DETACH); err != nil {
			log.Errorf("failed to unmount %s", err)
		}
	}
}

func updateHostname(hostname string) error {
	log.Infof("Set hostname to %s", hostname)

	// update /etc/hostname
	fHostname, err := os.OpenFile("/etc/hostname", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fHostname.Close()
	fmt.Fprint(fHostname, hostname)

	// update /etc/hosts
	fHosts, err := os.OpenFile("/etc/hosts", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fHosts.Close()
	fmt.Fprintf(fHosts, "127.0.0.1    %s.local %s\n", hostname, hostname)
	fmt.Fprint(fHosts, "127.0.0.1    localhost.localdomain localhost\n")

	return syscall.Sethostname([]byte(hostname))
}

func linkMtab() error {
	os.RemoveAll("/etc/mtab")
	return os.Symlink("../proc/self/mounts", "/etc/mtab")
}
