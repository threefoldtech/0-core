package bridge

import (
	"sync"

	"github.com/threefoldtech/0-core/base/pm"

	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/apps/plugins/nft"
	"github.com/threefoldtech/0-core/base/plugin"
)

var (
	log     = logging.MustGetLogger("bridge")
	manager Manager

	//Plugin plugin entry point
	Plugin = plugin.Plugin{
		Name:      "bridge",
		Version:   "1.0",
		CanUpdate: true,
		Requires:  []string{"nft"},
		Open: func(api plugin.API) (err error) {
			return initManager(&manager, api)
		},
		API: func() interface{} {
			return &manager
		},
		Actions: map[string]pm.Action{
			"create":     manager.create,
			"list":       manager.list,
			"delete":     manager.delete,
			"host-add":   manager.addHost,
			"nic-add":    manager.addNic,
			"nic-remove": manager.removeNic,
			"nic-list":   manager.listNic,
		},
	}
)

type Manager struct {
	api plugin.API
	m   sync.Mutex
}

func (m *Manager) nft() nft.API {
	return m.api.MustPlugin("nft").(nft.API)
}

func initManager(mgr *Manager, api plugin.API) error {
	mgr.api = api
	return nil
}
