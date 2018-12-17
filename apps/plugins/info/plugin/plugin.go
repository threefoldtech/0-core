package main

import (
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	//Plugin entry point
	Plugin = plugin.Plugin{
		Name:    "info",
		Version: "1.0",
		Actions: map[string]pm.Action{
			"cpu":     getCPUInfo,
			"mem":     getMemInfo,
			"disk":    getDiskInfo,
			"nic":     getNicInfo,
			"os":      getOsInfo,
			"port":    getPortInfo,
			"version": getVersionInfo,
		},
	}
)

func main() {} //silence errors
