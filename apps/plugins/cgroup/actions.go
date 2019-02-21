package cgroup

import (
	"encoding/json"

	"github.com/threefoldtech/0-core/base/pm"
)

//GroupArg basic cgroup arg
type GroupArg struct {
	Subsystem Subsystem `json:"subsystem"`
	Name      string    `json:"name"`
}

func (m *Manager) list(ctx pm.Context) (interface{}, error) {
	return m.GetGroups()
}

func (m *Manager) ensure(ctx pm.Context) (interface{}, error) {
	var args GroupArg
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	_, err := m.GetGroup(args.Subsystem, args.Name)

	return nil, err
}

func (m *Manager) remove(ctx pm.Context) (interface{}, error) {
	var args GroupArg
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	return nil, m.Remove(args.Subsystem, args.Name)
}

func (m *Manager) reset(ctx pm.Context) (interface{}, error) {
	var args GroupArg
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	group, err := m.Get(args.Subsystem, args.Name)
	if err != nil {
		return nil, pm.NotFoundError(err)
	}

	group.Reset()

	return nil, nil
}

func (m *Manager) tasks(ctx pm.Context) (interface{}, error) {
	var args GroupArg
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	group, err := m.Get(args.Subsystem, args.Name)
	if err != nil {
		return nil, pm.NotFoundError(err)
	}

	return group.Tasks()
}

func (m *Manager) taskAdd(ctx pm.Context) (interface{}, error) {
	var args struct {
		GroupArg
		PID int `json:"pid"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	group, err := m.Get(args.Subsystem, args.Name)
	if err != nil {
		return nil, pm.NotFoundError(err)
	}

	return nil, group.Task(args.PID)
}

func (m *Manager) taskRemove(ctx pm.Context) (interface{}, error) {
	var args struct {
		GroupArg
		PID int `json:"pid"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	group, err := m.Get(args.Subsystem, args.Name)
	if err != nil {
		return nil, pm.NotFoundError(err)
	}

	root := group.Root()
	return nil, root.Task(args.PID)
}

func (m *Manager) cpusetSpec(ctx pm.Context) (interface{}, error) {
	var args struct {
		Name string `json:"name,omitempty"`
		Cpus string `json:"cpus"`
		Mems string `json:"mems"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	group, err := m.Get(CPUSetSubsystem, args.Name)

	if err != nil {
		return nil, pm.NotFoundError(err)
	}

	if group, ok := group.(CPUSetGroup); ok {
		if len(args.Cpus) != 0 {
			if err := group.Cpus(args.Cpus); err != nil {
				return nil, err
			}
		}

		if len(args.Mems) != 0 {
			if err := group.Mems(args.Mems); err != nil {
				return nil, err
			}
		}

		args.Name = ""
		args.Cpus, _ = group.GetCpus()
		args.Mems, _ = group.GetMems()

		return args, nil
	}

	return nil, pm.InternalError(ErrInvalidType)
}

func (m *Manager) memorySpec(ctx pm.Context) (interface{}, error) {
	var args struct {
		Name string `json:"name,omitempty"`
		Mem  int    `json:"mem"`
		Sawp int    `json:"swap"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	group, err := m.Get(MemorySubsystem, args.Name)

	if err != nil {
		return nil, pm.NotFoundError(err)
	}

	if group, ok := group.(MemoryGroup); ok {
		if args.Mem != 0 {
			if err := group.Limit(args.Mem, args.Sawp); err != nil {
				return nil, err
			}
		}

		args.Name = ""
		mem, swap, _ := group.Limits()
		args.Mem = mem
		args.Sawp = swap - mem

		return args, nil
	}

	return nil, pm.InternalError(ErrInvalidType)
}
