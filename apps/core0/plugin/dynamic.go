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

const (
	PluginNamespace = "plugin"
)

var (
	log = logging.MustGetLogger("plugin")
)

//Plugin wrapper
type Plugin struct {
	*plg.Plugin
	IsOpen bool
}

//DynamicManager a plugin manager
type DynamicManager struct {
	BaseManager
	path []string
}

type ScopedManager struct {
	*BaseManager
	scope string
}

func (s ScopedManager) Store() plg.Store {
	return s.BaseManager.Store(s.scope)
}

//New create a new plugin manager
func NewDynamic(path ...string) (*DynamicManager, error) {
	m := &DynamicManager{
		BaseManager: newBaseManager(),
		path:        path,
	}

	return m, nil
}

//Get action from fqn
//implements Router
func (m *DynamicManager) Get(name string) (pm.Action, bool) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 0 {
		return nil, false
	}

	target := ""
	if len(parts) == 2 {
		target = parts[1]
	}

	if parts[0] == PluginNamespace {
		return m.internal(target)
	}

	return m.BaseManager.Get(name)
}

func (m *DynamicManager) internal(name string) (action pm.Action, ok bool) {
	switch name {
	case "list":
		action = m.list
	case "load":
		action = m.load
	}

	if action != nil {
		return action, true
	}

	return nil, false
}

func (m *DynamicManager) loadPlugins(p string) ([]*plg.Plugin, error) {
	log.Infof("loading plugin %s", p)
	plug, err := plugin.Open(p)
	if err != nil {
		return nil, err
	}

	sym, err := plug.Lookup("Plugin")
	if err != nil {
		return nil, err
	}

	switch sym := sym.(type) {
	case *plg.Plugin:
		return []*plg.Plugin{sym}, nil
	case *[]*plg.Plugin:
		return *sym, nil
	default:
		return nil, fmt.Errorf("plugin symbol of wrong type: %T", sym)
	}
}

func (m *DynamicManager) Load() error {
	for _, p := range m.path {
		if err := m.loadPath(p); err != nil {
			return err
		}
	}

	return m.BaseManager.Load()
}

func (m *DynamicManager) loadPath(p string) error {
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

		plugins, err := m.loadPlugins(path.Join(p, item.Name()))

		if err != nil {
			log.Errorf("failed to load '%s': %v", item.Name(), err)
			continue
		}
		m.l.Lock()
		for _, p := range plugins {
			m.plugins[p.Name] = &Plugin{Plugin: p}
		}
		m.l.Unlock()
	}

	return nil
}
