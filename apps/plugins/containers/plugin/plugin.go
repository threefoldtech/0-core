package main

import (
	"fmt"

	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/apps/plugins/cgroup"
	"github.com/threefoldtech/0-core/apps/plugins/socat"
	"github.com/threefoldtech/0-core/base/plugin"
)

var (
	log = logging.MustGetLogger("containers")

	manager containerManager

	Plugin = plugin.Plugin{
		Name:     "corex",
		Version:  "1.0",
		Requires: []string{"cgroup", "socat", "bridge"},
		Open: func(api plugin.API) error {
			return iniManager(&manager, api)
		},
	}
)

func main() {}

func iniManager(mgr *containerManager, api plugin.API) error {
	mgr.api = api
	var ok bool
	if api, err := api.Plugin("cgroup"); err != nil {
		if mgr.cgroup, ok = api.(cgroup.API); !ok {
			return fmt.Errorf("invalid cgroup api")
		}
	} else {
		return err
	}

	if api, err := api.Plugin("socat"); err != nil {
		if mgr.socat, ok = api.(socat.API); !ok {
			return fmt.Errorf("invalid socat api")
		}
	} else {
		return err
	}

	mgr.containers = make(map[uint16]*container)

	if err := mgr.setUpCGroups(); err != nil {
		return err
	}

	if err := mgr.setUpDefaultBridge(); err != nil {
		return err
	}

	return nil
}

/*
func ContainerSubsystem(sink *transport.Sink, cell *screen.RowCell) (containers.ContainerManager, error) {

	containerMgr := &containerManager{
		containers: make(map[uint16]*container),
		sink:       sink,
		cell:       cell,
	}

	cell.Text = "Containers: 0"

	if err := containerMgr.setUpCGroups(); err != nil {
		return nil, err
	}
	if err := containerMgr.setUpDefaultBridge(); err != nil {
		return nil, err
	}

	pm.RegisterBuiltIn(cmdContainerCreate, containerMgr.create)
	pm.RegisterBuiltIn(cmdContainerCreateSync, containerMgr.createSync)
	pm.RegisterBuiltIn(cmdContainerList, containerMgr.list)
	pm.RegisterBuiltIn(cmdContainerDispatch, containerMgr.dispatch)
	pm.RegisterBuiltIn(cmdContainerTerminate, containerMgr.terminate)
	pm.RegisterBuiltIn(cmdContainerFind, containerMgr.find)
	pm.RegisterBuiltIn(cmdContainerNicAdd, containerMgr.nicAdd)
	pm.RegisterBuiltIn(cmdContainerNicRemove, containerMgr.nicRemove)
	pm.RegisterBuiltIn(cmdContainerPortForwardAdd, containerMgr.portforwardAdd)
	pm.RegisterBuiltIn(cmdContainerPortForwardRemove, containerMgr.portforwardRemove)
	pm.RegisterBuiltIn(cmdContainerBackup, containerMgr.backup)
	pm.RegisterBuiltIn(cmdContainerRestore, containerMgr.restore)
	pm.RegisterBuiltIn(cmdContainerFListLayer, containerMgr.flistLayer)
	// flist specific commands
	pm.RegisterBuiltIn(cmdFlistCreate, containerMgr.flistCreate)

	//container specific info
	pm.RegisterBuiltIn(cmdContainerZerotierInfo, containerMgr.ztInfo)
	pm.RegisterBuiltIn(cmdContainerZerotierList, containerMgr.ztList)

	return containerMgr, nil
}
*/
