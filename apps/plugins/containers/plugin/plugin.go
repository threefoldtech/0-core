package main

import (
	"fmt"

	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/apps/plugins/cgroup"
	"github.com/threefoldtech/0-core/apps/plugins/socat"
	"github.com/threefoldtech/0-core/apps/plugins/zfs"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	log = logging.MustGetLogger("containers")

	manager containerManager

	Plugin = plugin.Plugin{
		Name:     "corex",
		Version:  "1.0",
		Requires: []string{"cgroup", "socat", "bridge", "zfs"},
		Open: func(api plugin.API) error {
			log.Debugf("initializing containers manager")
			return iniManager(&manager, api)
		},
		// pm.RegisterBuiltIn(cmdContainerCreate, containerMgr.create)
		// pm.RegisterBuiltIn(cmdContainerCreateSync, containerMgr.createSync)
		// pm.RegisterBuiltIn(cmdContainerList, containerMgr.list)
		// pm.RegisterBuiltIn(cmdContainerDispatch, containerMgr.dispatch)
		// pm.RegisterBuiltIn(cmdContainerTerminate, containerMgr.terminate)
		// pm.RegisterBuiltIn(cmdContainerFind, containerMgr.find)
		// pm.RegisterBuiltIn(cmdContainerNicAdd, containerMgr.nicAdd)
		// pm.RegisterBuiltIn(cmdContainerNicRemove, containerMgr.nicRemove)
		// pm.RegisterBuiltIn(cmdContainerPortForwardAdd, containerMgr.portforwardAdd)
		// pm.RegisterBuiltIn(cmdContainerPortForwardRemove, containerMgr.portforwardRemove)
		// pm.RegisterBuiltIn(cmdContainerBackup, containerMgr.backup)
		// pm.RegisterBuiltIn(cmdContainerRestore, containerMgr.restore)
		// pm.RegisterBuiltIn(cmdContainerFListLayer, containerMgr.flistLayer)
		// // flist specific commands
		// pm.RegisterBuiltIn(cmdFlistCreate, containerMgr.flistCreate)

		// //container specific info
		// pm.RegisterBuiltIn(cmdContainerZerotierInfo, containerMgr.ztInfo)
		// pm.RegisterBuiltIn(cmdContainerZerotierList, containerMgr.ztList)
		// cmdContainerDispatch          = "corex.dispatch"
		// cmdContainerFListLayer        = "corex.flist-layer"

		Actions: map[string]pm.Action{
			"create":             manager.create,
			"create-sync":        manager.createSync,
			"terminate":          manager.terminate,
			"list":               manager.list,
			"find":               manager.find,
			"nic-add":            manager.nicAdd,
			"nic-remove":         manager.nicRemove,
			"portforward-add":    manager.portforwardAdd,
			"portforward-remove": manager.portforwardRemove,
			"backup":             manager.backup,
			"restore":            manager.restore,
			"zerotier.inf":       manager.ztInfo,
			"zerotier.list":      manager.ztList,
			"flist-layer":        manager.flistLayer,
			"flist.create":       manager.flistCreate,
		},
	}
)

func main() {}

func iniManager(mgr *containerManager, api plugin.API) error {
	mgr.api = api
	var ok bool
	if api, err := api.Plugin("cgroup"); err == nil {
		if mgr.cgroup, ok = api.(cgroup.API); !ok {
			return fmt.Errorf("invalid cgroup api")
		}
	} else {
		return err
	}

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

	mgr.containers = make(map[uint16]*container)
	log.Debugf("setting up containers cgroups")
	if err := mgr.setUpCGroups(); err != nil {
		return err
	}

	log.Debugf("setting up containers default networking")
	if err := mgr.setUpDefaultBridge(); err != nil {
		return err
	}

	return nil
}
