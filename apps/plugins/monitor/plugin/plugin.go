package main

import (
	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	log     = logging.MustGetLogger("monitor")
	manager Manager

	Plugin = plugin.Plugin{
		Name:    "monitor",
		Version: "1.0",
		Open: func(api plugin.API) error {
			manager.api = api
			return nil
		},

		Actions: map[string]pm.Action{
			"cpu":     manager.monitor,
			"memory":  manager.monitor,
			"disk":    manager.monitor,
			"network": manager.monitor,
		},
	}
)

type Manager struct {
	api plugin.API
}

func main() {}
