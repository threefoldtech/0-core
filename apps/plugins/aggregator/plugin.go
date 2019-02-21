package aggregator

import (
	"time"

	"github.com/threefoldtech/0-core/base/pm"

	logging "github.com/op/go-logging"
	cache "github.com/patrickmn/go-cache"
	"github.com/threefoldtech/0-core/apps/plugins/protocol"
	"github.com/threefoldtech/0-core/base/plugin"
)

var (
	log     = logging.MustGetLogger("aggregator")
	manager Manager
	_       pm.StatsHandler = (*Manager)(nil)

	//Plugin entry point
	Plugin = plugin.Plugin{
		Name:      "aggregator",
		Version:   "1.0",
		CanUpdate: false,
		Requires:  []string{"protocol"},
		Open: func(api plugin.API) error {
			initManager(&manager, api)
			return nil
		},
		API: func() interface{} {
			return &manager
		},
		Actions: map[string]pm.Action{
			"query": manager.query,
		},
	}
)

func main() {}

type Manager struct {
	api   plugin.API
	cache *cache.Cache
}

func (m *Manager) database() protocol.Database {
	return m.api.MustPlugin("protocol").(protocol.API).Database()
}

func initManager(mgr *Manager, api plugin.API) error {
	mgr.api = api

	mgr.cache = cache.New(1*time.Hour, 5*time.Minute)
	mgr.cache.OnEvicted(func(key string, _ interface{}) {
		if _, err := mgr.database().DelKeys(key); err != nil {
			log.Errorf("failed to evict stats key %s", key)
		}
	})

	return nil
}
