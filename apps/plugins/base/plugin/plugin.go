package main

import (
	"github.com/threefoldtech/0-core/apps/plugins/base/core"
	"github.com/threefoldtech/0-core/apps/plugins/base/fs"
	"github.com/threefoldtech/0-core/apps/plugins/base/info"
	"github.com/threefoldtech/0-core/apps/plugins/base/ip"
	"github.com/threefoldtech/0-core/apps/plugins/base/job"
	"github.com/threefoldtech/0-core/base/plugin"
)

var (
	Plugin = []*plugin.Plugin{
		&core.Plugin,
		&fs.Plugin,
		&info.Plugin,
		&ip.Plugin,
		&job.Plugin,
	}
)

func main() {}
