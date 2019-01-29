package pprof

import (
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	manager Manager

	Plugin = plugin.Plugin{
		Name:      "pprof",
		Version:   "1.0",
		CanUpdate: true,
		Open: func(api plugin.API) error {
			manager.api = api
			return nil
		},

		Actions: map[string]pm.Action{
			"cpu.start": manager.pprofStart,
			"cpu.stop":  manager.pprofStop,
			"mem.write": manager.pprofMemWrite,
			"mem.stat":  manager.pprofMemStat,
		},
	}
)

type Manager struct {
	api plugin.API
}
