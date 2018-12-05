package mgr

import (
	"syscall"

	"github.com/threefoldtech/0-core/base/pm"
)

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
	pm.Process
	GetPID() int32
}

//ProcessFactory interface
type ProcessFactory func(PIDTable, *pm.Command) pm.Process
