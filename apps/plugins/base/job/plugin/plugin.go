package main

import (
	"github.com/threefoldtech/0-core/apps/plugins/job"
)

var (
	//Plugin is declared in the info package to support static building
	//against this plugin. Other plugins that are not needed to built
	//statically don't have to be done this way
	Plugin = job.Plugin
)

func main() {} //silence errors
