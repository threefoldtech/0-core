package builtin

import (
	"fmt"
	base "github.com/Zero-OS/0-Core/base"
	"github.com/Zero-OS/0-Core/base/pm"
	"github.com/Zero-OS/0-Core/base/pm/core"
	"github.com/Zero-OS/0-Core/base/pm/process"
)

const (
	cmdPing = "core.ping"
)

func init() {
	pm.CmdMap[cmdPing] = process.NewInternalProcessFactory(ping)
}

func ping(cmd *core.Command) (interface{}, error) {
	return fmt.Sprintf("PONG %s", base.Version()), nil
}
