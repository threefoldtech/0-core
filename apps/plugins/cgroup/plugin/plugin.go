package main

import (
	"os"
	"path"
	"syscall"

	"github.com/threefoldtech/0-core/apps/plugins/cgroup"
	"github.com/threefoldtech/0-core/base/pm"

	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/base/plugin"
)

var (
	//supported subsystems
	subsystems = map[cgroup.Subsystem]mkg{
		cgroup.DevicesSubsystem: mkDevicesGroup,
		cgroup.CPUSetSubsystem:  mkCPUSetGroup,
		cgroup.MemorySubsystem:  mkMemoryGroup,
	}

	log     = logging.MustGetLogger("cgroups")
	manager Manager

	_ cgroup.API = (*Manager)(nil)

	//Plugin export plugin
	Plugin = plugin.Plugin{
		Name:    "cgroup",
		Version: "1.0",
		Open: func(api plugin.API) error {
			return initManager(&manager, api)
		},
		API: func() interface{} {

			return &manager
		},
		Actions: map[string]pm.Action{
			"list":   manager.list,
			"ensure": manager.ensure,
			"remove": manager.remove,

			"tasks":       manager.tasks,
			"task-add":    manager.taskAdd,
			"task-remove": manager.taskRemove,

			"reset": manager.reset,

			"cpuset.spec": manager.cpusetSpec,
			"memory.spec": manager.memorySpec,
		},
	}
)

type Manager struct {
	api plugin.API
}

func initManager(m *Manager, api plugin.API) error {
	m.api = api
	os.MkdirAll(CGroupBase, 0755)
	err := syscall.Mount("cgroup_root", CGroupBase, "tmpfs", 0, "")
	if err != nil {
		return err
	}

	for sub := range subsystems {
		p := path.Join(CGroupBase, string(sub))
		os.MkdirAll(p, 0755)

		err = syscall.Mount(string(sub), p, "cgroup", 0, string(sub))
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {}
