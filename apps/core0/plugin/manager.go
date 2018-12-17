package plugin

import (
	"fmt"
	"io/ioutil"
	"path"
	"plugin"
	"strings"

	logging "github.com/op/go-logging"
	plg "github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	log = logging.MustGetLogger("pm")
)

//Manager a plugin manager
type Manager struct {
	path    []string
	plugins map[string]*plg.Plugin
}

//New create a new plugin manager
func New(path ...string) (*Manager, error) {
	m := &Manager{
		path:    path,
		plugins: make(map[string]*plg.Plugin),
	}

	return m, m.load()
}

//Get action from fqn
func (m *Manager) Get(name string) (pm.Action, bool) {
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

func (m *Manager) open(p string) (*plg.Plugin, error) {
	log.Infof("loading plugin %s", p)
	plug, err := plugin.Open(p)
	if err != nil {
		return nil, err
	}

	sym, err := plug.Lookup("Plugin")
	if err != nil {
		return nil, err
	}

	if plugin, ok := sym.(*plg.Plugin); ok {
		return plugin, nil
	}

	return nil, fmt.Errorf("plugin symbol of wrong type: %T", sym)
}

func (m *Manager) load() error {
	for _, p := range m.path {
		if err := m.loadPath(p); err != nil {
			return err
		}
	}

	for _, plugin := range m.plugins {
		if plugin.Open != nil {
			if err := plugin.Open(m); err != nil {
				log.Errorf("failed to initialize plugin %s: %s", plugin.Name, err)
				continue
			}
		}

		log.Infof("plugin %s loaded", plugin.Name)
	}

	return nil
}

func (m *Manager) loadPath(p string) error {
	items, err := ioutil.ReadDir(p)
	if err != nil {
		return err
	}

	for _, item := range items {
		if item.IsDir() {
			continue
		}

		if path.Ext(item.Name()) != ".so" {
			continue
		}

		plugin, err := m.open(path.Join(p, item.Name()))

		if err != nil {
			return err
		}

		m.plugins[plugin.Name] = plugin
	}

	return nil
}
