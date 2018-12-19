package main

import (
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	manager Manager

	Plugin = plugin.Plugin{
		Name:    "disk",
		Version: "1.0",
		Open: func(api plugin.API) error {
			manager.api = api
			return nil
		},

		// pm.RegisterBuiltIn("disk.getinfo", d.info)
		// pm.RegisterBuiltIn("disk.list", d.list)
		// pm.RegisterBuiltIn("disk.protect", d.protect)
		// pm.RegisterBuiltIn("disk.mounts", d.mounts)

		Actions: map[string]pm.Action{
			"getinfo":         manager.info,
			"list":            manager.list,
			"protect":         manager.protect,
			"mounts":          manager.mounts,
			"smartctl-info":   manager.smartctlInfo,
			"smartctl-health": manager.smartctlHealth,
			"spindown":        manager.spindown,
			"seektime":        manager.seektime,
		},
	}
)

type Manager struct {
	api plugin.API
}

func main() {}
