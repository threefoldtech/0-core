package main

import (
	"github.com/codegangsta/cli"
	"github.com/Zero-OS/0-Core/base/pm/core"
)

func reboot(t Transport, c *cli.Context) {
	response, err := t.Run(Command{
		Sync: true,
		Content: core.Command{
			Command: "core.reboot",
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	//you probably won't reach here but let's assume you did
	response.ValidateResultOrExit()
}

func poweroff(t Transport, c *cli.Context) {
	response, err := t.Run(Command{
		Sync: true,
		Content: core.Command{
			Command: "core.poweroff",
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	//you probably won't reach here but let's assume you did
	response.ValidateResultOrExit()
}
