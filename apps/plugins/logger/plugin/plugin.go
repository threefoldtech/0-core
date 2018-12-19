package main

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	"github.com/op/go-logging"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

var (
	Plugin = plugin.Plugin{
		Name:    "logger",
		Version: "1.0",

		Actions: map[string]pm.Action{
			"set_level": setLevel,
			"reopen":    reopen,
		},
	}
)

type LogLevel struct {
	Level string `json:"level"`
}

func setLevel(ctx pm.Context) (interface{}, error) {
	var args LogLevel
	cmd := ctx.Command()

	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	level, err := logging.LogLevel(args.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %s", args.Level)
	}

	logging.SetLevel(level, "")

	return nil, nil

}

func reopen(ctx pm.Context) (interface{}, error) {
	return nil, syscall.Kill(os.Getpid(), syscall.SIGUSR1)
}

func main() {}
