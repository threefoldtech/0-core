package main

import (
	"fmt"
	"time"

	libvirt "github.com/libvirt/libvirt-go"
	"github.com/threefoldtech/0-core/apps/plugins/containers"
	"github.com/threefoldtech/0-core/apps/plugins/socat"
	"github.com/threefoldtech/0-core/apps/plugins/zfs"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	manager kvmManager

	Plugin = plugin.Plugin{
		Name:      "kvm",
		Version:   "1.0",
		CanUpdate: false,
		Requires: []string{
			"bridge",
			"socat",
			"zfs",
			"corex",
		},
		Open: func(api plugin.API) error {
			log.Debugf("initializing containers manager")
			return iniManager(&manager, api)
		},

		Actions: map[string]pm.Action{
			kvmCreateCommand:            manager.create,
			kvmDestroyCommand:           manager.destroy,
			kvmShutdownCommand:          manager.shutdown,
			kvmRebootCommand:            manager.reboot,
			kvmResetCommand:             manager.reset,
			kvmPauseCommand:             manager.pause,
			kvmResumeCommand:            manager.resume,
			kvmInfoCommand:              manager.info,
			kvmInfoPSCommand:            manager.infops,
			kvmAttachDiskCommand:        manager.attachDisk,
			kvmDetachDiskCommand:        manager.detachDisk,
			kvmAddNicCommand:            manager.addNic,
			kvmRemoveNicCommand:         manager.removeNic,
			kvmLimitDiskIOCommand:       manager.limitDiskIO,
			kvmMigrateCommand:           manager.migrate,
			kvmListCommand:              manager.list,
			kvmPrepareMigrationTarget:   manager.prepareMigrationTarget,
			kvmCreateImage:              manager.createImage,
			kvmConvertImage:             manager.convertImage,
			kvmGetCommand:               manager.get,
			kvmPortForwardAddCommand:    manager.portforwardAdd,
			kvmPortForwardRemoveCommand: manager.portforwardRemove,
			//those next 2 commands should never be called by the client, unfortunately we don't have
			//support for internal commands yet.
			kvmMonitorCommand: manager.monitor,
			kvmEventsCommand:  manager.events,
		},
	}
)

func main() {}

func iniManager(mgr *kvmManager, api plugin.API) error {
	mgr.api = api
	mgr.evch = make(chan map[string]interface{}, 100)
	mgr.domainsInfo = make(map[string]*DomainInfo)
	mgr.devDeleteEvent = NewSync()

	var ok bool
	if api, err := api.Plugin("socat"); err == nil {
		if mgr.socat, ok = api.(socat.API); !ok {
			return fmt.Errorf("invalid socat api")
		}
	} else {
		return err
	}

	if api, err := api.Plugin("zfs"); err == nil {
		if mgr.filesystem, ok = api.(zfs.API); !ok {
			return fmt.Errorf("invalid zfs api")
		}
	} else {
		return err
	}

	if api, err := api.Plugin("corex"); err == nil {
		if mgr.container, ok = api.(containers.API); !ok {
			return fmt.Errorf("invalid containers api")
		}
	} else {
		return err
	}

	if err := libvirt.EventRegisterDefaultImpl(); err != nil {
		return err
	}

	go func() {
		for {
			if err := libvirt.EventRunDefaultImpl(); err != nil {
				log.Warningf("failed to register to kvm events, trying again ...")
				<-time.After(2 * time.Second)
			}
		}
	}()

	mgr.libvirt.lifeCycleHandler = mgr.domaineLifeCycleHandler
	mgr.libvirt.deviceRemovedHandler = mgr.deviceRemovedHandler
	mgr.libvirt.deviceRemovedFailedHandler = mgr.deviceRemovedFailedHandler

	if err := mgr.setupDefaultGateway(); err != nil {
		return err
	}
	//start domains monitoring command
	mgr.api.Run(&pm.Command{
		ID:              "kvm.monitor",
		Command:         "kvm.monitor",
		RecurringPeriod: 30,
	})

	//start events command
	mgr.api.Run(&pm.Command{
		ID:      "kvm.events",
		Command: "kvm.events",
	})

	return nil
}
