package main

import (
	"encoding/json"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/threefoldtech/0-core/base/pm"
)

func (d *Manager) pprofStart(ctx pm.Context) (interface{}, error) {
	var args struct {
		File string `json:"file"`
	}
	cmd := ctx.Command()

	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	fd, err := os.Create(args.File)
	if err != nil {
		return nil, err
	}

	return nil, pprof.StartCPUProfile(fd)
}

func (d *Manager) pprofStop(ctx pm.Context) (interface{}, error) {
	pprof.StopCPUProfile()
	return nil, nil
}

func (d *Manager) pprofMemWrite(ctx pm.Context) (interface{}, error) {
	var args struct {
		File string `json:"file"`
	}
	cmd := ctx.Command()

	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	fd, err := os.Create(args.File)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return nil, pprof.WriteHeapProfile(fd)
}

func (d *Manager) pprofMemStat(ctx pm.Context) (interface{}, error) {
	var stat runtime.MemStats
	runtime.ReadMemStats(&stat)
	return stat, nil
}
