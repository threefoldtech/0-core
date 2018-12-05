package mgr

import "github.com/threefoldtech/0-core/base/pm"

//Handler defines an interface to receiver the process manager events
//A handler can be any object that implements one or many handle methods below
type Handler interface{}

//ResultHandler receives the command result on exit
type ResultHandler interface {
	Result(cmd *pm.Command, result *pm.JobResult)
}

//MessageHandler gets called on the receive of each single message
//from all commands
type MessageHandler interface {
	Message(*pm.Command, *pm.Message)
}

//StatsHandler receives parsed stats messages
type StatsHandler interface {
	Stats(operation string, key string, value float64, id string, tags ...Tag)
}

//PreHandler is called with the commands before exectution
type PreHandler interface {
	Pre(cmd *pm.Command)
}
