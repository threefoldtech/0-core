package plugin

import (
	"strings"

	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

type StaticManager struct {
	BaseManager
}

//NewStatic creates a new static router
func NewStatic(pl ...*plugin.Plugin) (*StaticManager, error) {
	r := &StaticManager{
		BaseManager: newBaseManager(),
	}

	for _, p := range pl {
		r.plugins[p.Name] = &Plugin{Plugin: p, IsOpen: false}
	}

	return r, nil
}

//Get action from fqn
func (m *StaticManager) Get(name string) (pm.Action, bool) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 0 {
		return nil, false
	}
	plugin, ok := m.plugins[parts[0]]
	if !ok {
		return nil, false
	}

	target := ""
	if len(parts) == 2 {
		target = parts[1]
	}

	action, ok := plugin.Actions[target]
	return action, ok
}
