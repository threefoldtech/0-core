package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"syscall"

	"github.com/pborman/uuid"
	"github.com/threefoldtech/0-core/base/pm"
)

const cmdbin = "rtinfo-client"

type rtinfoParams struct {
	Host  string   `json:"host"`
	Port  uint     `json:"port"`
	Disks []string `json:"disks"`
	job   string
}

func (rtm *Manager) start(ctx pm.Context) (interface{}, error) {
	var args rtinfoParams
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	cmdargs := []string{
		"--host", args.Host, "--port", fmt.Sprintf("%d", args.Port),
	}

	for _, d := range args.Disks {
		cmdargs = append(cmdargs, "--disk", d)
	}

	rtm.m.Lock()
	defer rtm.m.Unlock()

	key := fmt.Sprintf("%s:%d", args.Host, args.Port)
	_, exists := rtm.info[key]
	if exists {
		return nil, pm.NotAcceptableError("an rtinfo agent running already for this daemon")
	}

	args.job = uuid.New()
	rtm.info[key] = &args

	rtinfocmd := &pm.Command{
		ID:      args.job,
		Command: pm.CommandSystem,
		Arguments: pm.MustArguments(
			pm.SystemCommandArguments{
				Name: cmdbin,
				Args: cmdargs,
			},
		),
	}

	onExit := &pm.ExitHook{
		Action: func(state bool) {
			rtm.m.Lock()
			delete(rtm.info, key)
			rtm.m.Unlock()
		},
	}

	_, err := manager.api.Run(rtinfocmd, onExit)

	if err != nil {
		//the process manager failed to start
		//hence we need to clean it up ourselvies
		rtm.m.Lock()
		delete(rtm.info, key)
		rtm.m.Unlock()
	}

	return nil, err
}

func (rtm *Manager) list(ctx pm.Context) (interface{}, error) {

	return rtm.info, nil
}

func (rtm *Manager) stop(ctx pm.Context) (interface{}, error) {
	var args struct {
		Host string `json:"host"`
		Port uint   `json:"port"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s:%d", args.Host, args.Port)
	rtm.m.RLock()
	defer rtm.m.RUnlock()
	info, exists := rtm.info[key]

	if !exists {
		return nil, nil
	}
	job, exists := rtm.api.JobOf(info.job)
	if exists {
		return job.Signal(syscall.SIGKILL), nil
	}

	return nil, errors.New("job wasn't running")
}
