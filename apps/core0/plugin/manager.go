package plugin

import (
	"fmt"
	"io/ioutil"
	"path"
	"plugin"
	"runtime/debug"
	"strings"

	"github.com/threefoldtech/0-core/base/mgr"

	logging "github.com/op/go-logging"
	plg "github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
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
}

//New create a new plugin manager
func New(path ...string) (*Manager, error) {
	m := &Manager{
		path:    path,
		plugins: make(map[string]*Plugin),
	}

	return m, nil
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

func (m *Manager) loadPlugin(p string) (*plg.Plugin, error) {
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

func (m *Manager) safeOpen(pl *Plugin) (err error) {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
			stack := debug.Stack()
			err = fmt.Errorf("paniced on plugin initialization: %v\n%s", e, string(stack))
		}
	}()

	err = pl.Open(m)
	return
}

func (m *Manager) openRecursive(pl *Plugin) error {
	if pl.IsOpen {
		return nil
	}

	for _, req := range pl.Requires {
		if _, ok := m.plugins[req]; !ok {
			return fmt.Errorf("plugin %s missing dep (%s)")
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

		plugin, err := m.loadPlugin(path.Join(p, item.Name()))

		if err != nil {
			return err
		}

		m.plugins[plugin.Name] = &Plugin{Plugin: plugin}
	}

	return nil
}
