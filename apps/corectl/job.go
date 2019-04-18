package main

import (
	"github.com/codegangsta/cli"
	"github.com/threefoldtech/0-core/base/pm"
)

func jobs(t Transport, c *cli.Context) {
	response, err := t.Run(Command{
		Sync: true,
		Content: pm.Command{
			Command:   "job.list",
			Arguments: pm.MustArguments(M{}),
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	response.ValidateResultOrExit()
	response.PrintYaml()
}

func jobKill(t Transport, c *cli.Context) {
	id := c.Args().First()
	if id == "" {
		log.Fatal("wrong usage")
	}

	response, err := t.Run(Command{
		Sync: true,
		Content: pm.Command{
			Command: "job.kill",
			Arguments: pm.MustArguments(M{
				"id": id,
			}),
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	response.ValidateResultOrExit()
}
