package main

import (
	"github.com/codegangsta/cli"
	client "github.com/threefoldtech/0-core/client/go-client"
)

func ping(t client.Client, c *cli.Context) {
	core := client.Core(t)
	PrintOrDie(core.Ping())
}
