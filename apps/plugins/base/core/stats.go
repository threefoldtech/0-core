package core

import (
	"github.com/threefoldtech/0-core/base/pm"
)

const (
	cmdGetAggregatedStats = "core.state"
)

func (mgr *coreManager) getStats(ctx pm.Context) (interface{}, error) {
	stat := pm.ProcessStats{}

	for _, runner := range mgr.api.Jobs() {
		ps := runner.Process()
		if ps == nil {
			continue
		}
		stats, ok := ps.(pm.Stater)
		if !ok {
			continue
		}

		processStats := stats.Stats()
		stat.CPU += processStats.CPU
		stat.RSS += processStats.RSS
		stat.Swap += processStats.Swap
		stat.VMS += processStats.VMS
	}

	//also get agent cpu and memory consumption.
	agentCPU, err := mgr.ps.Percent(0)
	if err == nil {
		stat.CPU += agentCPU
	} else {
		log.Errorf("%s", err)
	}

	agentMem, err := mgr.ps.MemoryInfo()
	if err == nil {
		stat.RSS += agentMem.RSS
		stat.Swap += agentMem.Swap
		stat.VMS += agentMem.VMS
	} else {
		log.Errorf("%s", err)
	}

	return stat, nil
}
