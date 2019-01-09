package cgroup

type Subsystem string

const (
	//DevicesSubsystem device subsystem
	DevicesSubsystem = Subsystem("devices")
	//CPUSetSubsystem cpu subsystem
	CPUSetSubsystem = Subsystem("cpuset")
	//MemorySubsystem memory subsystem
	MemorySubsystem = Subsystem("memory")
)

//Group generic cgroup interface
type Group interface {
	Name() string
	Subsystem() Subsystem
	Task(pid int) error
	Tasks() ([]int, error)
	Root() Group
	Reset()
}

//API defines cgroups plugin api
type API interface {
	GetGroup(subsystem Subsystem, name string) (Group, error)
	Get(subsystem Subsystem, name string) (Group, error)
	GetGroups() (map[Subsystem][]string, error)
	Remove(subsystem Subsystem, name string) error
	Exists(subsystem Subsystem, name string) bool
}

//DevicesGroup defines a device cgroup
type DevicesGroup interface {
	Group
	Deny(spec string) error
	Allow(spec string) error
	List() ([]string, error)
}

//MemoryGroup defines a memory group
type MemoryGroup interface {
	Group
	Limits() (int, int, error)
	Limit(mem, swap int) error
}

//CPUSetGroup defines a cpuset group
type CPUSetGroup interface {
	Group
	Cpus(sepc string) error
	Mems(sepc string) error
	GetCpus() (string, error)
	GetMems() (string, error)
}
