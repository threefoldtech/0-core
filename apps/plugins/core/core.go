package core

import (
	"os"

	logging "github.com/op/go-logging"

	psutil "github.com/shirou/gopsutil/process"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	log = logging.MustGetLogger("core")
	mgr coreManager

	Plugin = plugin.Plugin{
		Name:    "core",
		Version: "1.0",
		Open: func(api plugin.API) error {
			return initMgr(&mgr, api)
		},
		Actions: map[string]pm.Action{
			"ping":      mgr.ping,
			"subscribe": mgr.subscribe,
			"state":     mgr.getStats,
		},
	}
)

func initMgr(mgr *coreManager, api plugin.API) error {
	ps, err := psutil.NewProcess(int32(os.Getpid()))

	if err != nil {
		return err
	}

	mgr.api = api
	mgr.ps = ps
	return nil
}

type coreManager struct {
	api plugin.API
	ps  *psutil.Process
}
