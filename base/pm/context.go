package pm

import "github.com/threefoldtech/0-core/base/stream"

//Context defines execution context
type Context interface {
	Message(msg *stream.Message)
	Log(text string, level ...uint16)
	Command() *Command
}

//Action defines a module action end point
type Action func(Context) (interface{}, error)
