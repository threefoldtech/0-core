package mgr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"syscall"

	"github.com/threefoldtech/0-core/base/pm"
)

/*
Runnable represents a runnable built in function that can be managed by the process manager.
*/
type Runnable func(*pm.Command) (interface{}, error)
type RunnableWithCtx func(*Context) (interface{}, error)

type Context struct {
	Command *pm.Command

	ch chan *pm.Message
}

func (c *Context) Message(msg *pm.Message) {
	c.ch <- msg
}

func (c *Context) Log(text string, level ...uint16) {
	//optional level
	var l uint16 = pm.LevelStdout

	if len(level) == 1 {
		l = level[0]
	} else if len(level) > 1 {
		panic("only a single optional level is allowed")
	}

	c.Message(&pm.Message{
		Message: text,
		Meta:    pm.NewMeta(l),
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
func NewInternalProcess(runnable Runnable) pm.ProcessFactory {
	factory := func(_ pm.PIDTable, cmd *pm.Command) pm.Process {
		return &internalProcess{
			runnable: runnable,
			ctx: Context{
				Command: cmd,
			},
		}
	}

	return factory
}

func NewInternalProcessWithCtx(runnable RunnableWithCtx) pm.ProcessFactory {
	factory := func(_ pm.PIDTable, cmd *pm.Command) pm.Process {
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
func (process *internalProcess) Command() *pm.Command {
	return process.ctx.Command
}

/*
Run runs the internal process
*/
func (process *internalProcess) Run() (<-chan *pm.Message, error) {

	channel := make(chan *pm.Message)
	process.ctx.ch = channel

	go func(channel chan *pm.Message) {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("panic: %v", err)
				debug.PrintStack()
				m, _ := json.Marshal(fmt.Sprintf("%v", err))
				channel <- &pm.Message{
					Meta:    pm.NewMetaWithCode(http.StatusInternalServerError, pm.LevelResultJSON, pm.ExitErrorFlag),
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

		var msg *pm.Message

		if err != nil {
			var code uint32
			if err, ok := err.(pm.RunError); ok {
				code = uint32(err.Code())
			} else {
				code = http.StatusInternalServerError
			}

			m, _ := json.Marshal(err.Error())
			msg = &pm.Message{
				Meta:    pm.NewMetaWithCode(code, pm.LevelResultJSON, pm.ExitErrorFlag),
				Message: string(m),
			}
		} else {
			m, _ := json.Marshal(value)
			msg = &pm.Message{
				Meta:    pm.NewMeta(pm.LevelResultJSON, pm.ExitSuccessFlag),
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
