package mgr

import (
	"fmt"

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

	for _, router := range routers {
		action, ok := router.Get(cmd.Command)
		if ok {
			return NewInternalProcess(action), nil
		}
	}

	return nil, UnknownCommandErr
}

//RegisterExtension registers a new command (extension) so it can be executed via commands
func RegisterExtension(cmd string, exe string, workdir string, cmdargs []string, env map[string]string) error {
	if _, ok := factories[cmd]; ok {
		return fmt.Errorf("job factory with the same name already registered: %s", cmd)
	}

	factory := extensionProcessFactory(exe, workdir, cmdargs, env)
	factories[cmd] = factory
	return nil
}
