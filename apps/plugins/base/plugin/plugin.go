package main

import (
	"github.com/threefoldtech/0-core/apps/plugins/base/bridge"
	"github.com/threefoldtech/0-core/apps/plugins/base/btrfs"
	"github.com/threefoldtech/0-core/apps/plugins/base/config"
	"github.com/threefoldtech/0-core/apps/plugins/base/core"
	"github.com/threefoldtech/0-core/apps/plugins/base/disk"
	"github.com/threefoldtech/0-core/apps/plugins/base/fs"
	"github.com/threefoldtech/0-core/apps/plugins/base/info"
	"github.com/threefoldtech/0-core/apps/plugins/base/ip"
	"github.com/threefoldtech/0-core/apps/plugins/base/job"
	"github.com/threefoldtech/0-core/apps/plugins/base/monitor"
	"github.com/threefoldtech/0-core/apps/plugins/base/power"
	"github.com/threefoldtech/0-core/apps/plugins/base/pprof"
	"github.com/threefoldtech/0-core/apps/plugins/base/process"
	"github.com/threefoldtech/0-core/apps/plugins/base/web"
	"github.com/threefoldtech/0-core/base/plugin"
)

var (
	Plugin = []*plugin.Plugin{
		&core.Plugin,
		&fs.Plugin,
		&info.Plugin,
		&ip.Plugin,
		&job.Plugin,
		&bridge.Plugin,
		&btrfs.Plugin,
		&config.Plugin,
		&disk.Plugin,
		&power.Plugin,
		&pprof.Plugin,
		&monitor.Plugin,
		&web.Plugin,
		&process.Plugin,
	}
)

func main() {}
