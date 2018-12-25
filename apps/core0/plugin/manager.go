package plugin

import (
	"fmt"
	"io/ioutil"
	"path"
	"plugin"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/threefoldtech/0-core/base/mgr"

	logging "github.com/op/go-logging"
	plg "github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/utils"
)

var (
	log = logging.MustGetLogger("pm")
)

//Manager a plugin manager
type Manager struct {
	path    []string
	plugins map[string]*plg.Plugin
}

type plugins []*plg.Plugin

// Len is the number of elements in the collection.
func (l plugins) Len() int {
	return len(l)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (l plugins) Less(i, j int) bool {
	pi := l[i]
	pj := l[j]

	//any one with now requirements should come first
	if len(pi.Requires) == 0 {
		return true
	} else if len(pj.Requires) == 0 {
		return false
	}

	//if i is required by j, it should come first
	if utils.InString(pj.Requires, pi.Name) {
		return true
	}

	//other wise, j should come first
	return false
}

// Swap swaps the elements with indexes i and j.
func (l plugins) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

//New create a new plugin manager
func New(path ...string) (*Manager, error) {
	m := &Manager{
		path:    path,
		plugins: make(map[string]*plg.Plugin),
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

func (m *Manager) safeOpen(pl *plg.Plugin) (err error) {
	defer func() {
		if e := recover(); e != nil {
			stack := debug.Stack()
			err = fmt.Errorf("paniced on plugin initialization: %v\n%s", e, string(stack))
		}
	}()

	err = pl.Open(m)
	return
}

func (m *Manager) Load() error {
	for _, p := range m.path {
		if err := m.loadPath(p); err != nil {
			return err
		}
	}

	l := make(plugins, 0, len(m.plugins))

	for _, p := range m.plugins {
		l = append(l, p)
	}

	sort.Sort(l)
all:
	for _, plugin := range l {
		for _, req := range plugin.Requires {
			if _, ok := m.plugins[req]; !ok {
				log.Warning("plugin %s missing dep (%s) ... ignore", plugin, req)
				delete(m.plugins, req)
				continue all
			}
		}

		if plugin.Open != nil {
			if err := m.safeOpen(plugin); err != nil {
				log.Errorf("failed to initialize plugin %s: %s", plugin.Name, err)
				continue
			}
		}

		log.Infof("plugin %s loaded", plugin.Name)

		if plugin.API != nil {
			//if the plugin api implement any
			//of the handler interfaces, register it
			//as a handler in process manager
			api := plugin.API()
			switch api.(type) {
			case pm.MessageHandler:
			case pm.ResultHandler:
			case pm.PreHandler:
			default:
				continue
			}

			log.Infof("register %s as a handler", plugin)
			mgr.AddHandle(api)
		}
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

		m.plugins[plugin.Name] = plugin
	}

	return nil
}
