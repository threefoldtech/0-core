package main

import (
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

const (
	Bash = "/bin/sh"
)

var (
	api plugin.API

	Plugin = plugin.Plugin{
		Name:    "bash",
		Version: "1.0",
		Open: func(a plugin.API) error {
			api = a
			return nil
		},
		Actions: map[string]pm.Action{
			"": bash, //default action
		},
	}
)

func bash(ctx pm.Context) (interface{}, error) {
	var args struct {
		Script string `json:"script"`
	}

	api.System(
		Bash, "-c", args.Script,
	)

	return nil, nil
}
