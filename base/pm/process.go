package pm

import (
	"fmt"
	"io"
	"syscall"

	"github.com/threefoldtech/0-core/base/stream"
)

const (
	//CommandSystem is the first and built in `core.system` command
	CommandSystem = "core.system"
)

//SystemCommandArguments arguments to system command
type SystemCommandArguments struct {
	Name  string            `json:"name"`
	Dir   string            `json:"dir"`
	Args  []string          `json:"args"`
	Env   map[string]string `json:"env"`
	StdIn string            `json:"stdin"`
}

func (s *SystemCommandArguments) String() string {
	return fmt.Sprintf("%v %s %v (%s)", s.Env, s.Name, s.Args, s.Dir)
}

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
	Run() (<-chan *stream.Message, error)
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

//PIDer a process that can return a PID
type PIDer interface {
	Process
	GetPID() int32
}
