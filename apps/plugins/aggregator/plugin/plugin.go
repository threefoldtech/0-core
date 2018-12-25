package main

import (
	"fmt"
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
		Name:     "aggregator",
		Version:  "1.0",
		Requires: []string{"protocol"},
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
	protocol protocol.API
	cache    *cache.Cache
}

func initManager(mgr *Manager, api plugin.API) error {
	var ok bool
	if api, err := api.Plugin("protocol"); err == nil {
		if mgr.protocol, ok = api.(protocol.API); !ok {
			return fmt.Errorf("invalid protocol api")
		}
	} else {
		return err
	}

	mgr.cache = cache.New(1*time.Hour, 5*time.Minute)
	mgr.cache.OnEvicted(func(key string, _ interface{}) {
		if _, err := mgr.protocol.Database().DelKeys(key); err != nil {
			log.Errorf("failed to evict stats key %s", key)
		}
	})

	return nil
}
