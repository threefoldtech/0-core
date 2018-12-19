package main

import (
	"fmt"
	"strings"

	"github.com/threefoldtech/0-core/base"
	"github.com/threefoldtech/0-core/base/mgr"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

type Router struct {
	plugins map[string]*plugin.Plugin
}

func NewRouter(pl ...*plugin.Plugin) (*Router, error) {
	r := &Router{
		plugins: make(map[string]*plugin.Plugin),
	}
	for _, p := range pl {
		r.plugins[p.Name] = p
	}
	for _, p := range r.plugins {
		if p.Open == nil {
			continue
		}

		if err := p.Open(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

//Get action from fqn
func (m *Router) Get(name string) (pm.Action, bool) {
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

//Version return version of base core0
func (m *Router) Version() base.Ver {
	return base.Version()
}

//Run proxy function for mgr.Run
func (m *Router) Run(cmd *pm.Command, hooks ...pm.RunnerHook) (pm.Job, error) {
	return mgr.Run(cmd, hooks...)
}

//System proxy method for mgr.System
func (m *Router) System(bin string, args ...string) (*pm.JobResult, error) {
	return mgr.System(bin, args...)
}

//Internal proxy method for mgr.Internal
func (m *Router) Internal(cmd string, args pm.M, out interface{}) error {
	return mgr.Internal(cmd, args, out)
}

//JobOf proxy method for mgr.JobOf
func (m *Router) JobOf(id string) (pm.Job, bool) {
	return mgr.JobOf(id)
}

func (m *Router) Jobs() map[string]pm.Job {
	return mgr.Jobs()
}

//Plugin plugin API getter
func (m *Router) Plugin(name string) (interface{}, error) {
	plg, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin not found")
	}
	if plg.API == nil {
		return nil, fmt.Errorf("plugin does not define an API")
	}

	return plg.API(), nil
}

func (m *Router) Aggregate(op, key string, value float64, id string, tags ...pm.Tag) {
	mgr.Aggregate(op, key, value, id, tags...)
}

func (m *Router) Shutdown(except ...string) {
	mgr.Shutdown(except...)
}
