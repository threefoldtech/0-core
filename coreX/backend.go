package main

import (
	"github.com/g8os/core0/base/logger"
	"github.com/g8os/core0/base/pm/core"
	"github.com/g8os/core0/base/pm/stream"
	"github.com/op/go-logging"
)

type logBackend struct {
	logger logger.Logger
	cmd    core.Command
}

func (l *logBackend) Log(level logging.Level, _ int, r *logging.Record) error {
	l.logger.Log(&l.cmd, &stream.Message{
		Level:   1,
		Message: r.Message(),
		Epoch:   r.Time.Unix(),
	})

	return nil
}
