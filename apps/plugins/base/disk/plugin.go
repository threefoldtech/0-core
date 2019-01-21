package disk

import (
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	manager Manager

	Plugin = plugin.Plugin{
		Name:      "disk",
		Version:   "1.0",
		CanUpdate: true,
		Open: func(api plugin.API) error {
			manager.api = api
			return nil
		},
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
