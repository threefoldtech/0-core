package builtin

import (
	"github.com/Zero-OS/0-Core/base/pm"
	"github.com/Zero-OS/0-Core/base/pm/core"
	"github.com/Zero-OS/0-Core/base/pm/process"
	"github.com/Zero-OS/0-Core/base/settings"
)

type configMgr struct{}

func init() {
	c := (*configMgr)(nil)
	pm.CmdMap["config.get"] = process.NewInternalProcessFactory(c.get)
}

func (c *configMgr) get(cmd *core.Command) (interface{}, error) {
	return settings.Settings, nil
}
