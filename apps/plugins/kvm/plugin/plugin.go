package main

import (
	"fmt"

	"github.com/threefoldtech/0-core/apps/plugins/containers"
	"github.com/threefoldtech/0-core/apps/plugins/socat"
	"github.com/threefoldtech/0-core/apps/plugins/zfs"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	manager kvmManager

	Plugin = plugin.Plugin{
		Name:    "corex",
		Version: "1.0",
		Requires: []string{
			"socat",
			"zfs",
			"containers",
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

	// if api, err := api.Plugin("logger"); err == nil {
	// 	if mgr.logger, ok = api.(pm.MessageHandler); !ok {
	// 		return fmt.Errorf("invalid logger api")
	// 	}
	// } else {
	// 	return err
	// }

	if api, err := api.Plugin("containers"); err == nil {
		if mgr.container, ok = api.(containers.API); !ok {
			return fmt.Errorf("invalid containers api")
		}
	} else {
		return err
	}
	return nil
}
