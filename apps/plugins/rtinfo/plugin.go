package rtinfo

import (
	"sync"

	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	manager Manager
	Plugin  = plugin.Plugin{
		Name:      "rtinfo",
		Version:   "1.0",
		CanUpdate: false,
		Open: func(api plugin.API) error {
			manager.api = api
			manager.info = make(map[string]*rtinfoParams)
			return nil
		},
		Actions: map[string]pm.Action{
			"start": manager.start,
			"list":  manager.list,
			"stop":  manager.stop,
		},
	}
)

type Manager struct {
	api  plugin.API
	info map[string]*rtinfoParams
	m    sync.RWMutex
}

func main() {}
