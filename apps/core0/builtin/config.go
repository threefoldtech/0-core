package builtin

import (
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/settings"
)

type configMgr struct{}

func init() {
	c := (*configMgr)(nil)
	pm.RegisterBuiltIn("config.get", c.get)
}

func (c *configMgr) get(cmd *pm.Command) (interface{}, error) {
	return settings.Settings, nil
}
