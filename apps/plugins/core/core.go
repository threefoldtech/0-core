package core

import (
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	api plugin.API

	Plugin = plugin.Plugin{
		Name:    "core",
		Version: "1.0",
		Open: func(a plugin.API) error {
			api = a
			return nil
		},
		Actions: map[string]pm.Action{
			"ping":      ping,
			"subscribe": subscribe,
		},
	}
)
