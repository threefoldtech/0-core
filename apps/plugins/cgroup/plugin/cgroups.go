package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/threefoldtech/0-core/apps/plugins/cgroup"
)

type mkg func(name string, subsys cgroup.Subsystem) cgroup.Group

const (
	//CGroupBase base mount point
	CGroupBase = "/sys/fs/cgroup"
)

var (

	//ErrDoesNotExist does not exist error
	ErrDoesNotExist = fmt.Errorf("cgroup does not exist")
	//ErrInvalidType invalid cgroup type
	ErrInvalidType = fmt.Errorf("cgroup of invalid type")
)

// //Init Initialized the cgroup subsystem
// func Init() (err error) {
// 	once.Do(func() {

// 		// pm.RegisterBuiltIn("cgroup.list", list)
// 		// pm.RegisterBuiltIn("cgroup.ensure", ensure)
// 		// pm.RegisterBuiltIn("cgroup.remove", remove)

// 		// pm.RegisterBuiltIn("cgroup.tasks", tasks)
// 		// pm.RegisterBuiltIn("cgroup.task-add", taskAdd)
// 		// pm.RegisterBuiltIn("cgroup.task-remove", taskRemove)

// 		// pm.RegisterBuiltIn("cgroup.reset", reset)

// 		// pm.RegisterBuiltIn("cgroup.cpuset.spec", cpusetSpec)
// 		// pm.RegisterBuiltIn("cgroup.memory.spec", memorySpec)
// 	})

// 	return
// }

//GetGroup creaes a group if it does not exist
func (m *Manager) GetGroup(subsystem cgroup.Subsystem, name string) (cgroup.Group, error) {
	mkg, ok := subsystems[subsystem]
	if !ok {
		return nil, fmt.Errorf("unknown subsystem '%s'", subsystem)
	}

	p := path.Join(CGroupBase, string(subsystem), name)
	if _, err := os.Stat(p); err == nil {
		//group was created before
		return mkg(name, subsystem), nil
	}

	if err := os.Mkdir(p, 0755); err != nil {
		return nil, err
	}

	group := mkg(name, subsystem)
	group.Reset()
	return group, nil
}

//Get group only if it exists
func (m *Manager) Get(subsystem cgroup.Subsystem, name string) (cgroup.Group, error) {
	if !m.Exists(subsystem, name) {
		return nil, ErrDoesNotExist
	}

	return m.GetGroup(subsystem, name)
}

//GetGroups gets all the available groups names grouped by susbsytem
func (m *Manager) GetGroups() (map[cgroup.Subsystem][]string, error) {
	result := make(map[cgroup.Subsystem][]string)
	for sub := range subsystems {
		// skip devices subsystem (only cpuset and memory)
		if sub == cgroup.DevicesSubsystem {
			continue
		}
		info, err := ioutil.ReadDir(path.Join(CGroupBase, string(sub)))
		if err != nil {
			return nil, err
		}
		for _, dir := range info {
			if !dir.IsDir() {
				continue
			}
			result[sub] = append(result[sub], dir.Name())
		}
	}

	return result, nil
}

//Remove removes a cgroup
func (m *Manager) Remove(subsystem cgroup.Subsystem, name string) error {
	if !m.Exists(subsystem, name) {
		return nil
	}

	builder := subsystems[subsystem]
	group := builder(name, subsystem)
	tasks, err := group.Tasks()
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		return os.Remove(path.Join(CGroupBase, string(subsystem), name))
	}

	root := group.Root()
	for _, task := range tasks {
		root.Task(task)
	}

	return os.Remove(path.Join(CGroupBase, string(subsystem), name))
}

//Exists Check if a cgroup exists
func (m *Manager) Exists(subsystem cgroup.Subsystem, name string) bool {
	_, ok := subsystems[subsystem]
	if !ok {
		return false
	}

	p := path.Join(CGroupBase, string(subsystem), name)
	info, err := os.Stat(p)
	if err != nil {
		return false
	}

	return info.IsDir()
}

type Group struct {
	name   string
	subsys cgroup.Subsystem
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Subsystem() cgroup.Subsystem {
	return g.subsys
}

func (g *Group) base() string {
	return path.Join(CGroupBase, string(g.subsys), g.name)
}

func (g *Group) Task(pid int) error {
	return ioutil.WriteFile(path.Join(g.base(), "cgroup.procs"), []byte(fmt.Sprint(pid)), 0644)
}

func (g *Group) Tasks() ([]int, error) {
	raw, err := ioutil.ReadFile(path.Join(g.base(), "cgroup.procs"))
	if err != nil {
		return nil, err
	}

	pids := make([]int, 0)
	for _, s := range strings.Split(string(raw), "\n") {
		if len(s) == 0 {
			continue
		}
		var pid int
		if _, err := fmt.Sscanf(s, "%d", &pid); err != nil {
			return nil, err
		}
		pids = append(pids, pid)
	}

	return pids, nil
}
