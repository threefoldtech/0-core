package pm

import (
	"io"
	"syscall"
)

const (
	//CommandSystem is the first and built in `core.system` command
	CommandSystem = "core.system"
)

//ProcessStats holds process cpu and memory usage
type ProcessStats struct {
	CPU  float64 `json:"cpu"`
	RSS  uint64  `json:"rss"`
	VMS  uint64  `json:"vms"`
	Swap uint64  `json:"swap"`
}

//Process interface
type Process interface {
	Command() *Command
	Run() (<-chan *Message, error)
}

//Channel is a 2 way communication channel that is mainly used
//to talk to the main containerd process `coreX`
type Channel interface {
	io.ReadWriteCloser
}

//ContainerProcess interface
type ContainerProcess interface {
	Process
	Channel() Channel
}

//Signaler a process that supports signals
type Signaler interface {
	Process
	Signal(sig syscall.Signal) error
}

//Stater a process that supports stats query
type Stater interface {
	Process
	Stats() *ProcessStats
}

//GetPID returns a PID of a process
type GetPID func() (int, error)

//PIDTable a table that keeps track of running process ids
type PIDTable interface {
	//PIDTable atomic registration of PID. MUST grantee that that no wait4 will happen
	//on any of the child process until the register operation is done.
	RegisterPID(g GetPID) (int, error)
	//WaitPID waits for a certain ID until it exits
	WaitPID(pid int) syscall.WaitStatus
}

//PIDer a process that can return a PID
type PIDer interface {
	Process
	GetPID() int32
}

//ProcessFactory interface
type ProcessFactory func(PIDTable, *Command) Process
