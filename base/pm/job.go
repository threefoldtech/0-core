package pm

import (
	"syscall"

	"github.com/threefoldtech/0-core/base/stream"
)

const (
	StandardStreamBufferSize = 100 //buffer size for each of stdout and stderr
	GenericStreamBufferSize  = 10  //we only keep last 100 message of all types.
)

type Job interface {
	Command() *Command
	Signal(sig syscall.Signal) error
	Process() Process
	Wait() *JobResult
	StartTime() int64
	Subscribe(stream.MessageHandler)
	Unschedule()
}
