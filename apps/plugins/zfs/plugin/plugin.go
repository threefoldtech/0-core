package main

import (
	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/apps/plugins/zfs"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	log     = logging.MustGetLogger("zfs")
	manager Manager
	_       zfs.API = (*Manager)(nil)

	//Plugin plugin entry point
	Plugin = plugin.Plugin{
		Name:      "zfs",
		Version:   "1.0",
		CanUpdate: true,
		Open: func(api plugin.API) error {
			manager.api = api
			return nil
		},
		API: func() interface{} {
			return &manager
		},
		Actions: map[string]pm.Action{
			"mount": manager.mount,
		},
	}
)

func main() {}

//Manager struct
type Manager struct {
	api plugin.API
}
