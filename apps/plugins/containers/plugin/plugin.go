package main

import (
	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/apps/plugins/containers"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	log = logging.MustGetLogger("containers")

	manager Manager
	_       containers.API = (*Manager)(nil)

	Plugin = plugin.Plugin{
		Name:      "corex",
		Version:   "1.0",
		CanUpdate: false,
		Requires: []string{
			"cgroup",
			"socat",
			"bridge",
			"zfs",
			"protocol",
			"logger",
		},
		API: func() interface{} {
			return &manager
		},
		Open: func(api plugin.API) error {
			log.Debugf("initializing containers manager")
			return iniManager(&manager, api)
		},

		Actions: map[string]pm.Action{
			"create":             manager.create,
			"create-sync":        manager.createSync,
			"terminate":          manager.terminate,
			"list":               manager.list,
			"find":               manager.find,
			"dispatch":           manager.dispatch,
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

func iniManager(mgr *Manager, api plugin.API) error {
	mgr.api = api

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
