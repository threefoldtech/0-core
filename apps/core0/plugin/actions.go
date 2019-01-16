package plugin

import (
	"encoding/json"

	"github.com/threefoldtech/0-core/base/pm"
)

func (m *Manager) list(ctx pm.Context) (interface{}, error) {
	m.l.RLock()
	defer m.l.RUnlock()

	var l []string
	for _, p := range m.plugins {
		l = append(l, p.String())
	}

	return l, nil
}

func (m *Manager) load(ctx pm.Context) (interface{}, error) {
	var args struct {
		Path string `json:"path"`
	}

	if err := json.Unmarshal(*ctx.Command().Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	pl, err := m.loadPlugin(args.Path)
	if err != nil {
		return nil, err
	}
	m.l.Lock()
	defer m.l.Unlock()
	plugin := &Plugin{Plugin: pl}
	if err := m.openRecursive(plugin); err != nil {
		return nil, err
	}

	m.plugins[pl.Name] = plugin

	return nil, nil
}
