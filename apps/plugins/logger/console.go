package logger

import (
	"github.com/op/go-logging"
	"github.com/threefoldtech/0-core/base/stream"
)

var (
	log = logging.MustGetLogger("logger")
)

type LogRecord struct {
	Core    uint16          `json:"core"`
	Command string          `json:"command"`
	Message *stream.Message `json:"message"`
}

func IsLoggable(defaults []uint16, msg *stream.Message) bool {
	if len(defaults) > 0 {
		return msg.Meta.Assert(defaults...)
	}

	return true
}

// ConsoleLogger Message message to the console
type ConsoleLogger struct {
	defaults []uint16
}

// newConsoleLogger creates a simple console logger that prints Message messages to Console.
func newConsoleLogger(defaults []uint16) Logger {
	return &ConsoleLogger{
		defaults: defaults,
	}
}

func (logger *ConsoleLogger) LogRecord(record *LogRecord) {
	if !IsLoggable(logger.defaults, record.Message) {
		return
	}

	if len(record.Message.Message) > 0 {
		log.Debugf("[%d]%s %s", record.Core, record.Command, record.Message)
	}
}
