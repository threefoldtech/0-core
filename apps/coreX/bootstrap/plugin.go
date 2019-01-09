package bootstrap

import (
	"fmt"
	"strings"

	"github.com/threefoldtech/0-core/base/mgr"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/utils"
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

func (b *Bootstrap) plugin(domain string, plugin Plugin) {
	if plugin.Queue {
		//if plugin requires queuing we make sure when a command is pushed (from a cient)
		//that we force a domain on it.
		mgr.AddHandle(&pluginPreProcessor{domain})
	}

	for _, export := range plugin.Exports {
		cmd := fmt.Sprintf("%s.%s", domain, export)
		mgr.RegisterExtension(cmd, plugin.Path, "", []string{export, "{}"}, nil)
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
