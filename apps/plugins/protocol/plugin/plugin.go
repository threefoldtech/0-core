package main

import (
	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/apps/plugins/protocol"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	log     = logging.MustGetLogger("transport")
	manager Manager
	_       protocol.API     = (*Manager)(nil)
	_       pm.ResultHandler = (*Manager)(nil)

	//Plugin entry point
	Plugin = plugin.Plugin{
		Name:      "protocol",
		Version:   "1.0",
		CanUpdate: false,
		//we require nft just to make sure firewall rules are applied before accepting connections
		Requires: []string{"nft"},
		Open: func(api plugin.API) error {
			return initManager(&manager, api)
		},
		API: func() interface{} {
			return &manager
		},
	}
)

func main() {}

func initManager(mgr *Manager, api plugin.API) error {
	pool := newPool()
	mgr.api = api
	mgr.pool = pool

	go mgr.process()

	return nil
}
