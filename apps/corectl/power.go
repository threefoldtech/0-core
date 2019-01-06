package main

import (
	"github.com/codegangsta/cli"
	client "github.com/threefoldtech/0-core/client/go-client"
)

func reboot(t client.Client, c *cli.Context) {
	power := client.Power(t)
	if err := power.Reboot(); err != nil {
		log.Fatal(err)
	}
}

func poweroff(t client.Client, c *cli.Context) {
	power := client.Power(t)
	if err := power.PowerOff(); err != nil {
		log.Fatal(err)
	}
}
