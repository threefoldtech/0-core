package plugin

import (
	"fmt"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/threefoldtech/0-core/base/mgr"
	plg "github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

//BaseManager implements basic plugin manager functionality
type BaseManager struct {
	plugins map[string]*Plugin

	l sync.RWMutex

	stores map[string]plg.Store
	sm     sync.Mutex
}

func newBaseManager() BaseManager {
	return BaseManager{
		plugins: make(map[string]*Plugin),
		stores:  make(map[string]plg.Store),
	}
}

//Get action from fqn
//implements Router
func (m *BaseManager) Get(name string) (pm.Action, bool) {
	parts := strings.SplitN(name, ".", 2)
	if len(parts) == 0 {
		return nil, false
	}

	target := ""
	if len(parts) == 2 {
		target = parts[1]
	}

	m.l.RLock()
	plugin, ok := m.plugins[parts[0]]
	m.l.RUnlock()

	if !ok {
		return nil, false
	}

	action, ok := plugin.Actions[target]
	return action, ok
}

func (m *BaseManager) safeOpen(pl *Plugin) (err error) {
	defer func() {
		if e := recover(); e != nil {
			debug.PrintStack()
			stack := debug.Stack()
			err = fmt.Errorf("paniced on plugin initialization: %v\n%s", e, string(stack))
		}
	}()

	err = pl.Open(ScopedManager{BaseManager: m, scope: pl.Name})
	return
}

func (m *BaseManager) openRecursive(pl *Plugin) error {
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

func (m *BaseManager) Load() error {
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
	m.l.Lock()
	for _, bad := range errored {
		delete(m.plugins, bad)
	}
	m.l.Unlock()

	return nil
}
