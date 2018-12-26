package main

import (
	"fmt"

	"github.com/threefoldtech/0-core/apps/plugins/protocol"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/settings"
	"github.com/threefoldtech/0-core/base/stream"
)

var (
	manager Manager

	//Plugin entry point
	Plugin = plugin.Plugin{
		Name:     "logger",
		Version:  "1.0",
		Requires: []string{"protocol"},
		Open: func(api plugin.API) error {
			return initManager(&manager, api)
		},
		API: func() interface{} {
			return &manager
		},
		Actions: map[string]pm.Action{
			"set_level": setLevel,
			"reopen":    reopen,
		},
	}
)

func main() {}

// Logger interface
type Logger interface {
	LogRecord(record *LogRecord)
}

//Manager struct
type Manager struct {
	api    plugin.API
	logger []Logger
}

func initManager(mgr *Manager, api plugin.API) error {
	var db protocol.Database
	if api, err := api.Plugin("protocol"); err == nil {
		var ok bool
		if db, ok = api.(protocol.Database); !ok {
			return fmt.Errorf("invalid database interface")
		}
	} else {
		return err
	}

	mgr.api = api
	mgr.logger = append(mgr.logger,
		newConsoleLogger(settings.Settings.Logging.File.Levels),
		//newTrunkLogger(db, settings.Settings.Logging.Ledis.Levels, settings.Settings.Logging.Ledis.Size),
		newStreamLogger(db, 0),
	)

	return nil
}

//Message handler implementation
func (m *Manager) Message(cmd *pm.Command, msg *stream.Message) {
	m.logRecord(&LogRecord{
		Command: cmd.ID,
		Message: msg,
	})
}

func (m *Manager) logRecord(record *LogRecord) {
	for _, logger := range m.logger {
		logger.LogRecord(record)
	}
}
