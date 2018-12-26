package main

import (
	"fmt"

	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/apps/plugins/cgroup"
	"github.com/threefoldtech/0-core/apps/plugins/protocol"
	"github.com/threefoldtech/0-core/apps/plugins/socat"
	"github.com/threefoldtech/0-core/apps/plugins/zfs"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	log = logging.MustGetLogger("containers")

	manager Manager

	Plugin = plugin.Plugin{
		Name:    "corex",
		Version: "1.0",
		Requires: []string{
			"cgroup",
			"socat",
			"bridge",
			"zfs",
			"protocol",
			"logger",
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

	if api, err := api.Plugin("protocol"); err == nil {
		if mgr.protocol, ok = api.(protocol.API); !ok {
			return fmt.Errorf("invalid protocol api")
		}
	} else {
		return err
	}

	if api, err := api.Plugin("logger"); err == nil {
		if mgr.logger, ok = api.(pm.MessageHandler); !ok {
			return fmt.Errorf("invalid logger api")
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
