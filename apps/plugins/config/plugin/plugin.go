package main

import (
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/settings"
)

var (
	Plugin = plugin.Plugin{
		Name:    "config",
		Version: "1.0",

		Actions: map[string]pm.Action{
			"get": get,
		},
	}
)

func get(ctx pm.Context) (interface{}, error) {
	return settings.Settings, nil
}

func main() {}
