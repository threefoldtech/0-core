package main

import (
	"github.com/threefoldtech/0-core/apps/core0/plugin"
	"github.com/threefoldtech/0-core/apps/plugins/aggregator"
	"github.com/threefoldtech/0-core/apps/plugins/base"
	"github.com/threefoldtech/0-core/apps/plugins/cgroup"
	"github.com/threefoldtech/0-core/apps/plugins/containers"
	"github.com/threefoldtech/0-core/apps/plugins/kvm"
	"github.com/threefoldtech/0-core/apps/plugins/logger"
	"github.com/threefoldtech/0-core/apps/plugins/nft"
	"github.com/threefoldtech/0-core/apps/plugins/protocol"
	"github.com/threefoldtech/0-core/apps/plugins/rtinfo"
	"github.com/threefoldtech/0-core/apps/plugins/socat"
	"github.com/threefoldtech/0-core/apps/plugins/zfs"
)

//GetPluginsManager returns a static plugin manager where plugins
//are built statically into core0 binary, no dynamic loading is supported
func GetPluginsManager() (*plugin.StaticManager, error) {
	plugins := base.Plugin
	plugins = append(
		plugins,
		&aggregator.Plugin,
		&cgroup.Plugin,
		&containers.Plugin,
		&kvm.Plugin,
		&logger.Plugin,
		&nft.Plugin,
		&protocol.Plugin,
		&rtinfo.Plugin,
		&socat.Plugin,
		&zfs.Plugin,
	)

	return plugin.NewStatic(plugins...)
}
