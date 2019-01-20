package main

import (
	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	log     = logging.MustGetLogger("power")
	manager Manager

	Plugin = plugin.Plugin{
		Name:      "power",
		Version:   "1.0",
		CanUpdate: true,
		Open: func(api plugin.API) error {
			manager.api = api
			return nil
		},

		Actions: map[string]pm.Action{
			"reboot":   manager.restart,
			"poweroff": manager.poweroff,
			"update":   manager.update,
		},
	}
)

type Manager struct {
	api plugin.API
}

func main() {}
