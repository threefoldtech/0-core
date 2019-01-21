package plugin

import (
	"fmt"
	"io/ioutil"
	"path"
	"plugin"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/threefoldtech/0-core/base/mgr"

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

//Manager a plugin manager
type Manager struct {
	path    []string
	plugins map[string]*Plugin

	l sync.RWMutex

	stores map[string]plg.Store
	sm     sync.Mutex
}

type ScopedManager struct {
	*Manager
	scope string
}

func (s ScopedManager) Store() plg.Store {
	return s.Manager.Store(s.scope)
}

//New create a new plugin manager
func New(path ...string) (*Manager, error) {
	m := &Manager{
		path:    path,
		plugins: make(map[string]*Plugin),
		stores:  make(map[string]plg.Store),
	}

	return m, nil
}

func (m *Manager) internal(name string) (action pm.Action, ok bool) {
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

//Get action from fqn
//implements Router
func (m *Manager) Get(name string) (pm.Action, bool) {
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

	plugin, ok := m.plugins[parts[0]]
	if !ok {
		return nil, false
	}

	action, ok := plugin.Actions[target]
	return action, ok
}

func (m *Manager) loadPlugins(p string) ([]*plg.Plugin, error) {
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

func (m *Manager) safeOpen(pl *Plugin) (err error) {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
			stack := debug.Stack()
			err = fmt.Errorf("paniced on plugin initialization: %v\n%s", e, string(stack))
		}
	}()

	err = pl.Open(ScopedManager{Manager: m, scope: pl.Name})
	return
}

func (m *Manager) openRecursive(pl *Plugin) error {
	if pl.IsOpen {
		return nil
	}

	for _, req := range pl.Requires {
		if _, ok := m.plugins[req]; !ok {
			return fmt.Errorf("plugin %s missing dep (%s)", pl.Name, req)
		}

		if err := m.openRecursive(m.plugins[req]); err != nil {
			return err
		}
	}

	if pl.Open != nil {
		if err := m.safeOpen(pl); err != nil {
			return err
		}
	}

	pl.IsOpen = true
	log.Infof("plugin %s loaded", pl.Name)

	return nil
}

func (m *Manager) Load() error {
	m.l.Lock()
	defer m.l.Unlock()

	for _, p := range m.path {
		if err := m.loadPath(p); err != nil {
			return err
		}
	}

	var errored []string

	for _, plugin := range m.plugins {
		if err := m.openRecursive(plugin); err != nil {
			log.Errorf("failed to initialize plugin %s: %s", plugin.Name, err)
			errored = append(errored, plugin.Name)
			continue
		}

		if plugin.API != nil {
			//if the plugin api implement any
			//of the handler interfaces, register it
			//as a handler in process manager
			api := plugin.API()
			switch api.(type) {
			case pm.MessageHandler:
			case pm.ResultHandler:
			case pm.PreHandler:
			case pm.StatsHandler:
			default:
				continue
			}

			log.Infof("register %s as a handler", plugin)
			mgr.AddHandle(api)
		}
	}

	for _, bad := range errored {
		delete(m.plugins, bad)
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

		plugins, err := m.loadPlugins(path.Join(p, item.Name()))

		if err != nil {
			log.Errorf("failed to load '%s': %v", item.Name(), err)
			continue
		}

		for _, p := range plugins {
			m.plugins[p.Name] = &Plugin{Plugin: p}
		}
	}

	return nil
}
