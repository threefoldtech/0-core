package plugin

import (
	"fmt"
	"io/ioutil"
	"path"
	"plugin"

	logging "github.com/op/go-logging"
	plg "github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	log = logging.MustGetLogger("pm")
)

//Manager a plugin manager
type Manager struct {
	path    string
	plugins map[string]*plg.Plugin
}

//New create a new plugin manager
func New(path string) (*Manager, error) {
	m := &Manager{
		path:    path,
		plugins: make(map[string]*plg.Plugin),
	}

	return m, m.load()
}

//Get action from fqn
func (m *Manager) Get(name string) (pm.Action, error) {
	return nil, nil
}

func (m *Manager) open(p string) (*plg.Plugin, error) {
	log.Info("loading plugin %s", p)
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
	items, err := ioutil.ReadDir(m.path)
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

		plugin, err := m.open(path.Join(m.path, item.Name()))

		if err != nil {
			return err
		}

		if plugin.Open != nil {
			if err := plugin.Open(m); err != nil {
				log.Errorf("failed to initialize plugin %s: %s", plugin.Name, err)
				continue
			}
		}

		log.Infof("plugin %s loaded", plugin.Name)
		m.plugins[plugin.Name] = plugin
	}

	return nil
}
