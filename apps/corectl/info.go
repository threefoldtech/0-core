package main

import (
	"github.com/codegangsta/cli"
	client "github.com/threefoldtech/0-core/client/go-client"
)

func info_cpu(t client.Client, c *cli.Context) {
	info := client.Info(t)
	PrintOrDie(info.CPU())
}

func info_disk(t client.Client, c *cli.Context) {
	info := client.Info(t)
	PrintOrDie(info.Disk())
}

func info_mem(t client.Client, c *cli.Context) {
	info := client.Info(t)
	PrintOrDie(info.Mem())
}

func info_nic(t client.Client, c *cli.Context) {
	info := client.Info(t)
	PrintOrDie(info.Nic())
}

func info_os(t client.Client, c *cli.Context) {
	info := client.Info(t)
	PrintOrDie(info.OS())
}
