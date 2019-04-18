package bootstrap

import (
	"fmt"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/utils"
	"strings"
)

const (
	PluginSearchPath = "/var/lib/corex/plugins"
	ManifestSymbol   = "Manifest"
	PluginSymbol     = "Plugin"
	PluginExt        = ".so"
)

type Plugin struct {
	Path    string
	Queue   bool
	Exports []string
}

type PluginsSettings struct {
	Plugin map[string]Plugin
}

type pluginPreProcessor struct {
	domain string
}

func (p *pluginPreProcessor) Pre(cmd *pm.Command) {
	if strings.HasPrefix(cmd.Command, p.domain+".") {
		cmd.Queue = p.domain
	}
}

func (b *Bootstrap) pluginFactory(plugin *Plugin, fn string) pm.ProcessFactory {
	return func(table pm.PIDTable, srcCmd *pm.Command) pm.Process {
		cmd := &pm.Command{
			ID: srcCmd.ID,
			Arguments: pm.MustArguments(pm.SystemCommandArguments{
				Name: plugin.Path,
				Args: []string{fn, string(*srcCmd.Arguments)},
			}),
			Queue:           srcCmd.Queue,
			StatsInterval:   srcCmd.StatsInterval,
			MaxTime:         srcCmd.MaxTime,
			MaxRestart:      srcCmd.MaxRestart,
			RecurringPeriod: srcCmd.RecurringPeriod,
			LogLevels:       srcCmd.LogLevels,
			Tags:            srcCmd.Tags,
		}

		return pm.NewSystemProcess(table, cmd)
	}
}

func (b *Bootstrap) plugin(domain string, plugin Plugin) {
	if plugin.Queue {
		//if plugin requires queuing we make sure when a command is pushed (from a cient)
		//that we force a domain on it.
		pm.AddHandle(&pluginPreProcessor{domain})
	}

	for _, export := range plugin.Exports {
		cmd := fmt.Sprintf("%s.%s", domain, export)
		pm.Register(cmd, b.pluginFactory(&plugin, export))
	}
}

func (b *Bootstrap) plugins() error {
	var plugins PluginsSettings
	if err := utils.LoadTomlFile("/.plugin.toml", &plugins); err != nil {
		return err
	}

	for domain, plugin := range plugins.Plugin {
		b.plugin(domain, plugin)
	}

	return nil
}
