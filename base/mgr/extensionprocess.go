package mgr

import (
	"encoding/json"
	"fmt"
	"syscall"

	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/stream"
	"github.com/threefoldtech/0-core/base/utils"
)

type extensionProcess struct {
	system pm.Process
	cmd    *pm.Command
}

func extensionProcessFactory(exe string, dir string, args []string, env map[string]string) ProcessFactory {
	constructor := func(table PIDTable, cmd *pm.Command) pm.Process {
		sysargs := pm.SystemCommandArguments{
			Name: exe,
			Dir:  dir,
			Env:  env,
		}

		var input map[string]interface{}
		if err := json.Unmarshal(*cmd.Arguments, &input); err != nil {
			log.Errorf("Failed to load extension command arguments: %s", err)
		}

		if stdin, ok := input["stdin"]; ok {
			switch in := stdin.(type) {
			case string:
				sysargs.StdIn = in
			case []byte:
				sysargs.StdIn = string(in)
			default:
				log.Errorf("invalid stdin to extension command, expecting string, or bytes")
			}

			delete(input, "stdin")
		}

		for _, arg := range args {
			if arg == "{}" {
				//global replace for the entire arguments string as is
				sysargs.Args = append(sysargs.Args, string(*cmd.Arguments))
			} else {
				sysargs.Args = append(sysargs.Args, utils.Format(arg, input))
			}
		}

		extcmd := &pm.Command{
			ID:        cmd.ID,
			Command:   pm.CommandSystem,
			Arguments: pm.MustArguments(sysargs),
			Tags:      cmd.Tags,
		}

		return &extensionProcess{
			system: newSystemProcess(table, extcmd),
			cmd:    cmd,
		}
	}

	return constructor
}

func (process *extensionProcess) Command() *pm.Command {
	return process.cmd
}

func (process *extensionProcess) Run() (<-chan *stream.Message, error) {
	return process.system.Run()
}

func (process *extensionProcess) Signal(sig syscall.Signal) error {
	if ps, ok := process.system.(pm.Signaler); ok {
		return ps.Signal(sig)
	}

	return fmt.Errorf("not supported")
}

func (process *extensionProcess) Stats() *pm.ProcessStats {
	if sys, ok := process.system.(pm.Stater); ok {
		return sys.Stats()
	}

	return nil
}
