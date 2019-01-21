package web

import (
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	manager Manager

	Plugin = plugin.Plugin{
		Name:      "web",
		Version:   "1.0",
		CanUpdate: true,
		Open: func(api plugin.API) error {
			manager.api = api
			return nil
		},

		Actions: map[string]pm.Action{
			"download": manager.downloadCmd,
		},
	}
)

type Manager struct {
	api plugin.API
}
