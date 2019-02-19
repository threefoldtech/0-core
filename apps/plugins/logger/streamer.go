package logger

import (
	"encoding/json"
	"fmt"

	"github.com/threefoldtech/0-core/apps/plugins/protocol"

	"github.com/threefoldtech/0-core/base/stream"
)

const (
	MaxStreamRedisQueueSize = 1000
	StreamRedisQueueTTL     = 60
)

// redisLogger send Message to redis queue
type streamLogger struct {
	db   protocol.Database
	size int64

	ch chan *LogRecord
}

// NewRedisLogger creates new redis logger handler
func newStreamLogger(db protocol.Database, size int64) Logger {
	if size == 0 {
		size = MaxStreamRedisQueueSize
	}

	rl := &streamLogger{
		db:   db,
		size: size,
		ch:   make(chan *LogRecord, MaxStreamRedisQueueSize),
	}

	go rl.pusher()
	return rl
}

func (l *streamLogger) LogRecord(record *LogRecord) {
	if !record.Message.Meta.Is(stream.StreamFlag) {
		//stream flag is not set
		return
	}

	l.ch <- record
}

func (l *streamLogger) pusher() {
	for {
		if err := l.push(); err != nil {
			//we don't sleep to avoid blocking the logging channel and to not slow down processes.
		}
	}
}

func (l *streamLogger) push() error {
	for {
		record := <-l.ch
		bytes, err := json.Marshal(record)
		if err != nil {
			continue
		}

		queue := fmt.Sprintf("stream:%s", record.Command)
		if _, err := l.db.RPush(queue, bytes); err != nil {
			return err
		}

		if err := l.db.LTrim(queue, -1*l.size, -1); err != nil {
			return err
		}

		l.db.LExpire(queue, StreamRedisQueueTTL)
	}
}
