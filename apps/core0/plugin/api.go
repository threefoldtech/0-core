package plugin

import (
	"fmt"

	"github.com/threefoldtech/0-core/base/mgr"
	"github.com/threefoldtech/0-core/base/pm"
)

//Run proxy function for mgr.Run
func (m *Manager) Run(cmd *pm.Command, hooks ...pm.RunnerHook) (pm.Job, error) {
	return mgr.Run(cmd, hooks...)
}

//System proxy method for mgr.System
func (m *Manager) System(bin string, args ...string) (*pm.JobResult, error) {
	return mgr.System(bin, args...)
}

//Internal proxy method for mgr.Internal
func (m *Manager) Internal(cmd string, args pm.M, out interface{}) error {
	return mgr.Internal(cmd, args, out)
}

//JobOf proxy method for mgr.JobOf
func (m *Manager) JobOf(id string) (pm.Job, bool) {
	return mgr.JobOf(id)
}

//Plugin plugin API getter
func (m *Manager) Plugin(name string) (interface{}, error) {
	plg, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found")
	}
	if plg.API == nil {
		return nil, fmt.Errorf("plugin does not define an API")
	}

	return plg.API(), nil
}
