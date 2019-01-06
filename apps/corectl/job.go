package main

import (
	"github.com/codegangsta/cli"
	client "github.com/threefoldtech/0-core/client/go-client"
)

func jobs(t client.Client, c *cli.Context) {
	core := client.Core(t)
	PrintOrDie(core.Jobs())
}

func jobKill(t client.Client, c *cli.Context) {
	core := client.Core(t)
	id := c.Args().First()
	if id == "" {
		log.Fatal("wrong usage")
	}

	if err := core.KillJob(client.JobId(id), 0); err != nil {
		log.Fatal(err)
	}
}
