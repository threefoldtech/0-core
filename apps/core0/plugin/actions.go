package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/threefoldtech/0-core/base/pm"
)

type info struct {
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Updatable bool     `json:"updateable"`
	Requires  []string `json:"requires,omitempty"`
}

func (m *Manager) list(ctx pm.Context) (interface{}, error) {
	m.l.RLock()
	defer m.l.RUnlock()

	var l []info
	for _, p := range m.plugins {
		l = append(l, info{
			Name:      p.Name,
			Version:   p.Version,
			Updatable: p.CanUpdate,
			Requires:  p.Requires,
		})
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

	pls, err := m.loadPlugins(args.Path)
	if err != nil {
		return nil, err
	}

	m.l.Lock()
	defer m.l.Unlock()
	for _, pl := range pls {
		if current, ok := m.plugins[pl.Name]; ok {
			if !current.CanUpdate {
				return nil, fmt.Errorf("plugin %s does not support hot update", current.Name)
			}
		}

		plugin := &Plugin{Plugin: pl}
		if err := m.openRecursive(plugin); err != nil {
			return nil, err
		}

		m.plugins[pl.Name] = plugin
	}

	return nil, nil
}
