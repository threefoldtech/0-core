package main

import (
	"github.com/codegangsta/cli"
	client "github.com/threefoldtech/0-core/client/go-client"
)

func system(t client.Client, c *cli.Context) {
	if c.Args().First() == "" {
		log.Fatalf("missing command to execute")
		return
	}
	sync := !c.GlobalBool("async")

	core := client.Core(t)
	jobID, err := core.SystemArgs(c.Args().First(), c.Args().Tail(), nil, "", "")

	if err != nil {
		log.Fatal(err)
	}

	if !sync {
		log.Infof("Job started with ID: %s", jobID)
		return
	}

	PrintResultOrDie(t.Result(jobID, c.GlobalInt("timeout")))
}
