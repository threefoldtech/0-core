package job

import (
	"encoding/json"
	"fmt"
	"syscall"

	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

const (
	cmdJobList       = "job.list"
	cmdJobKill       = "job.kill"
	cmdJobKillAll    = "job.killall"
	cmdJobUnschedule = "job.unschedule"
)

var (
	api plugin.API
	//Plugin entry point
	Plugin = plugin.Plugin{
		Name:      "job",
		Version:   "1.0",
		CanUpdate: true,
		Open: func(a plugin.API) error {
			api = a
			return nil
		},
		Actions: map[string]pm.Action{
			"list":       list,
			"kill":       kill,
			"unschedule": unschedule,
		},
	}
)

type jobArguments struct {
	ID string `json:"id"`
}

type processData struct {
	pm.ProcessStats
	StartTime int64       `json:"starttime"`
	Cmd       *pm.Command `json:"cmd,omitempty"`
	PID       int32       `json:"pid"`
}

func list(ctx pm.Context) (interface{}, error) {
	//load data
	var data jobArguments
	cmd := ctx.Command()
	err := json.Unmarshal(*cmd.Arguments, &data)
	if err != nil {
		return nil, err
	}

	var stats []processData
	var runners []pm.Job

	if data.ID != "" {
		job, ok := api.JobOf(data.ID)

		if !ok {
			return nil, fmt.Errorf("Process with id '%s' doesn't exist", data.ID)
		}

		runners = []pm.Job{job}
	} else {
		for _, runner := range api.Jobs() {
			runners = append(runners, runner)
		}
	}

	for _, runner := range runners {
		s := processData{
			Cmd:       runner.Command(),
			StartTime: runner.StartTime(),
		}

		ps := runner.Process()

		if stater, ok := ps.(pm.Stater); ok {
			psStat := stater.Stats()
			s.CPU = psStat.CPU
			s.RSS = psStat.RSS
			s.VMS = psStat.VMS
			s.Swap = psStat.Swap
		}

		if pider, ok := ps.(pm.PIDer); ok {
			s.PID = pider.GetPID()
		}

		stats = append(stats, s)
	}

	return stats, nil
}

type jobKillArguments struct {
	jobArguments
	Signal syscall.Signal `json:"signal"`
}

func kill(ctx pm.Context) (interface{}, error) {
	//load data
	data := jobKillArguments{}
	cmd := ctx.Command()
	err := json.Unmarshal(*cmd.Arguments, &data)

	if err != nil {
		return nil, err
	}

	if data.Signal == syscall.Signal(0) {
		data.Signal = syscall.SIGTERM
	}

	job, ok := api.JobOf(data.ID)
	if !ok {
		return false, nil
	}

	if err := job.Signal(data.Signal); err != nil {
		return false, err
	}

	return true, nil

}

func unschedule(ctx pm.Context) (interface{}, error) {
	//load data
	data := jobArguments{}
	cmd := ctx.Command()
	err := json.Unmarshal(*cmd.Arguments, &data)

	if err != nil {
		return nil, err
	}

	job, ok := api.JobOf(data.ID)
	if !ok {
		return false, nil
	}

	job.Unschedule()

	return true, nil
}
