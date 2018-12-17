package kvm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"syscall"

	"github.com/threefoldtech/0-core/apps/core0/helper/filesystem"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/settings"
	"github.com/threefoldtech/0-core/base/utils"
	yaml "gopkg.in/yaml.v2"
)

func (m *kvmManager) flistConfigOverride(target string, cfg map[string]string) error {
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

func (m *kvmManager) flistMount(uuid, src, storage string, cfg map[string]string) (config FListBootConfig, err error) {
	namespace := fmt.Sprintf(VmNamespaceFmt, uuid)

	if storage == "" {
		storage = settings.Settings.Globals.Get("storage", "zdb://hub.grid.tf:9900")
	}

	target := path.Join(VmBaseRoot, uuid)
	onExit := &pm.ExitHook{
		Action: func(e bool) {
			//destroy this machine, if fs exited
			conn, err := m.libvirt.getConnection()
			if err != nil {
				log.Errorf("failed to get libvirt connection: %s", err)
				return
			}
			domain, err := conn.LookupDomainByUUIDString(uuid)
			if err != nil {
				return
			}
			log.Warningf("VM (%s) filesystem exited while running, destorying the machine", uuid)
			m.destroyDomain(uuid, domain)
		},
	}

	if err = filesystem.MountFList(namespace, storage, src, target, onExit); err != nil {
		return
	}

	//make sure that root filesystem has those dirs. mostly required for
	//the machine to boot. Also the hub deletes the dev directory
	for _, d := range []string{"dev", "sys", "proc"} {
		os.MkdirAll(path.Join(target, d), 0755)
	}

	defer func() {
		if err != nil {
			m.flistUnmount(uuid)
		}
	}()

	if err = m.flistConfigOverride(target, cfg); err != nil {
		return
	}

	//load entry config
	cfgstr, err := ioutil.ReadFile(path.Join(target, "boot", "boot.yaml"))
	if err != nil {
		return config, fmt.Errorf("failed to open boot/boot.yaml: %s", err)
	}

	err = yaml.Unmarshal(cfgstr, &config)
	config.Root = target
	return
}

func (m *kvmManager) flistUnmount(uuid string) error {
	target := path.Join(VmBaseRoot, uuid)
	err := syscall.Unmount(target, syscall.MNT_FORCE)
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			if errno == syscall.EINVAL {
				return nil
			}
		}
		return err
	}

	os.RemoveAll(target)
	os.RemoveAll(fmt.Sprintf(VmNamespaceFmt, uuid))
	return nil
}
