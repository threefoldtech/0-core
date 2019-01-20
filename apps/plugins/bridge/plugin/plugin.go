package main

import (
	"fmt"
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

func main() {}

type Manager struct {
	api plugin.API
	nft nft.API
	m   sync.Mutex
}

func initManager(mgr *Manager, api plugin.API) error {
	mgr.api = api
	nftPlugin, err := api.Plugin("nft")
	if err != nil {
		return err
	}

	if nftApi, ok := nftPlugin.(nft.API); ok {
		mgr.nft = nftApi
	} else {
		return fmt.Errorf("invalid nft interface")
	}

	return nil
}
