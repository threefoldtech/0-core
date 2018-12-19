package main

import (
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	manager Manager

	Plugin = plugin.Plugin{
		Name:    "btrfs",
		Version: "1.0",
		Open: func(api plugin.API) error {
			manager.api = api
			return nil
		},

		Actions: map[string]pm.Action{
			"list":            manager.List,
			"info":            manager.Info,
			"create":          manager.Create,
			"device_add":      manager.DeviceAdd,
			"device_remove":   manager.DeviceRemove,
			"subvol_create":   manager.SubvolCreate,
			"subvol_quota":    manager.SubvolQuota,
			"subvol_list":     manager.SubvolList,
			"subvol_snapshot": manager.SubvolSnapshot,
		},
	}
)

type Manager struct {
	api plugin.API
}

func main() {}
