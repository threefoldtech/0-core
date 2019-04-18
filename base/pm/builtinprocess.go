package pm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"syscall"

	"github.com/threefoldtech/0-core/base/pm/stream"
)

/*
Runnable represents a runnable built in function that can be managed by the process manager.
*/
type Runnable func(*Command) (interface{}, error)
type RunnableWithCtx func(*Context) (interface{}, error)

type Context struct {
	Command *Command

	ch chan *stream.Message
}

func (c *Context) Message(msg *stream.Message) {
	c.ch <- msg
}

func (c *Context) Log(text string, level ...uint16) {
	//optional level
	var l uint16 = stream.LevelStdout

	if len(level) == 1 {
		l = level[0]
	} else if len(level) > 1 {
		panic("only a single optional level is allowed")
	}

	c.Message(&stream.Message{
		Message: text,
		Meta:    stream.NewMeta(l),
	})
}

/*
internalProcess implements a Procss interface and represents an internal (go) process that can be managed by the process manager
*/
type internalProcess struct {
	runnable interface{}
	ctx      Context
}

/*
NewInternalProcess factory to build Runnable processes
*/
func NewInternalProcess(runnable Runnable) ProcessFactory {
	factory := func(_ PIDTable, cmd *Command) Process {
		return &internalProcess{
			runnable: runnable,
			ctx: Context{
				Command: cmd,
			},
		}
	}

	return factory
}

func NewInternalProcessWithCtx(runnable RunnableWithCtx) ProcessFactory {
	factory := func(_ PIDTable, cmd *Command) Process {
		return &internalProcess{
			runnable: runnable,
			ctx: Context{
				Command: cmd,
			},
		}
	}

	return factory
}

/*
Cmd returns the internal process command
*/
func (process *internalProcess) Command() *Command {
	return process.ctx.Command
}

/*
Run runs the internal process
*/
func (process *internalProcess) Run() (<-chan *stream.Message, error) {

	channel := make(chan *stream.Message)
	process.ctx.ch = channel

	go func(channel chan *stream.Message) {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("panic: %v", err)
				debug.PrintStack()
				m, _ := json.Marshal(fmt.Sprintf("%v", err))
				channel <- &stream.Message{
					Meta:    stream.NewMetaWithCode(http.StatusInternalServerError, stream.LevelResultJSON, stream.ExitErrorFlag),
					Message: string(m),
				}
			}

			close(channel)
		}()

		var value interface{}
		var err error
		switch runnable := process.runnable.(type) {
		case Runnable:
			value, err = runnable(process.ctx.Command)
		case RunnableWithCtx:
			value, err = runnable(&process.ctx)
		}

		var msg *stream.Message

		if err != nil {
			var code uint32
			if err, ok := err.(RunError); ok {
				code = uint32(err.Code())
			} else {
				code = http.StatusInternalServerError
			}

			m, _ := json.Marshal(err.Error())
			msg = &stream.Message{
				Meta:    stream.NewMetaWithCode(code, stream.LevelResultJSON, stream.ExitErrorFlag),
				Message: string(m),
			}
		} else {
			m, _ := json.Marshal(value)
			msg = &stream.Message{
				Meta:    stream.NewMeta(stream.LevelResultJSON, stream.ExitSuccessFlag),
				Message: string(m),
			}
		}

		channel <- msg
	}(channel)

	return channel, nil
}

func (process *internalProcess) Signal(sig syscall.Signal) error {
	return nil
}
