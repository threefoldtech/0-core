package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/base/pm"
)

func setLevel(ctx pm.Context) (interface{}, error) {
	var args struct {
		Level string `json:"level"`
	}
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
