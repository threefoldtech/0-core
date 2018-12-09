package mgr

import (
	"github.com/threefoldtech/0-core/base/pm"
)

//implement internal processes

/*
Global command ProcessConstructor registery
*/
var factories = map[string]ProcessFactory{
	pm.CommandSystem: NewSystemProcess,
}

//GetProcessFactory gets a process factory from command name
func getFactory(cmd *pm.Command) (ProcessFactory, error) {
	factory, ok := factories[cmd.Command]
	if ok {
		return factory, nil
	}

	if router == nil {
		return nil, UnknownCommandErr
	}

	action, err := router.Get(cmd.Command)
	if err != nil {
		return nil, err
	}

	return NewInternalProcess(action), nil
}
