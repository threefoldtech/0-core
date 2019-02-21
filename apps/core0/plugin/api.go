package plugin

import (
	"fmt"

	"github.com/threefoldtech/0-core/base"
	"github.com/threefoldtech/0-core/base/mgr"
	plg "github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

//Version return version of base core0
func (m *BaseManager) Version() base.Ver {
	return base.Version()
}

//Run proxy function for mgr.Run
func (m *BaseManager) Run(cmd *pm.Command, hooks ...pm.RunnerHook) (pm.Job, error) {
	return mgr.Run(cmd, hooks...)
}

//System proxy method for mgr.System
func (m *BaseManager) System(bin string, args ...string) (*pm.JobResult, error) {
	return mgr.System(bin, args...)
}

//Internal proxy method for mgr.Internal
func (m *BaseManager) Internal(cmd string, args pm.M, out interface{}) error {
	return mgr.Internal(cmd, args, out)
}

//JobOf proxy method for mgr.JobOf
func (m *BaseManager) JobOf(id string) (pm.Job, bool) {
	return mgr.JobOf(id)
}

func (m *BaseManager) Jobs() map[string]pm.Job {
	return mgr.Jobs()
}

//Plugin plugin API getter
func (m *BaseManager) Plugin(name string) (interface{}, error) {
	m.l.RLock()
	defer m.l.RUnlock()
	plg, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found")
	}
	if plg.API == nil {
		return nil, fmt.Errorf("plugin does not define an API")
	}

	return plg.API(), nil
}

func (m *BaseManager) MustPlugin(name string) interface{} {
	plugin, err := m.Plugin(name)
	if err != nil {
		panic(fmt.Sprintf("plugin %v not found", name))
	}

	return plugin
}

func (m *BaseManager) Shutdown(except ...string) {
	mgr.Shutdown(except...)
}

func (m *BaseManager) Aggregate(op, key string, value float64, id string, tags ...pm.Tag) {
	mgr.Aggregate(op, key, value, id, tags...)
}

func (m *BaseManager) Store(scope string) plg.Store {
	m.sm.Lock()
	defer m.sm.Unlock()

	store, ok := m.stores[scope]
	if ok {
		return store
	}

	store = newStore()
	m.stores[scope] = store

	return store
}
