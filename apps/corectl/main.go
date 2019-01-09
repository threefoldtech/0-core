package main

import (
	"os"
	"path"

	"github.com/codegangsta/cli"
	"github.com/op/go-logging"
)

var (
	log = logging.MustGetLogger("corectl")
)

func init() {
	formatter := logging.MustStringFormatter("%{color}%{message}%{color:reset}")
	logging.SetFormatter(formatter)
}

func main() {
	app := cli.NewApp()
	app.Name = "corectl"
	app.Usage = "manage g8os"
	app.UsageText = "Query or send commands to g8os manager"
	app.Version = "1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "socket, s",
			Value: "unix:///var/run/redis.sock",
			Usage: "Path to core socket",
		},
		cli.IntFlag{
			Name:  "timeout, t",
			Value: 0,
			Usage: "Commands that takes longer than this will get killed",
		},
		cli.BoolFlag{
			Name:  "async",
			Usage: "Run command asyncthronuslly (only commands that supports this)",
		},
		cli.StringFlag{
			Name:  "container",
			Usage: "Container numeric ID or comma seperated list with tags (only with execute)",
		},
		cli.StringFlag{
			Name:  "id",
			Usage: "Speicify porcess id, if not given a random guid will be generated",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "ping",
			Usage:  "checks connectivity with g8os",
			Action: WithClient(ping),
		},
		{
			Name:            "execute",
			Usage:           "execute arbitary commands",
			Description:     "execute any command inside the core context",
			ArgsUsage:       "command [args]",
			Action:          WithClient(system),
			SkipFlagParsing: true,
		},
		{
			Name:      "stop",
			Usage:     "stops a process with `id`",
			ArgsUsage: "id",
			Action:    WithClient(jobKill),
		},
		{
			Name:   "job",
			Usage:  "control system jobs",
			Action: WithClient(jobs),
			Subcommands: []cli.Command{
				{
					Name:   "list",
					Action: WithClient(jobs),
					Usage:  "list jobs",
				},
				{
					Name:      "kill",
					Action:    WithClient(jobKill),
					Usage:     "kill a job with `id`",
					ArgsUsage: "id",
				},
			},
		},
		{
			Name:   "container",
			Usage:  "container management",
			Action: WithClient(containers),
			Subcommands: []cli.Command{
				{
					Name:      "list",
					Action:    WithClient(containers),
					Usage:     "list containers (default)",
					ArgsUsage: "tag [tag...]",
				},
				{
					Name:      "inspect",
					Action:    WithClient(containerInspect),
					Usage:     "print detailed container info",
					ArgsUsage: "id",
				},
			},
		},
		{
			Name:  "info",
			Usage: "query various infomation",
			Subcommands: []cli.Command{
				{
					Name:   "cpu",
					Usage:  "display CPU info",
					Action: WithClient(info_cpu),
				},
				{
					Name:    "memory",
					Aliases: []string{"mem"},
					Usage:   "display memory info",
					Action:  WithClient(info_mem),
				},
				{
					Name:   "disk",
					Usage:  "display disk info",
					Action: WithClient(info_disk),
				},
				{
					Name:   "nic",
					Usage:  "display NIC info",
					Action: WithClient(info_nic),
				},
				{
					Name:   "os",
					Usage:  "display OS info",
					Action: WithClient(info_os),
				},
			},
		},
		{
			Name:   "reboot",
			Usage:  "reboot the machine",
			Action: WithClient(reboot),
		},
		{
			Name:   "poweroff",
			Usage:  "Shuts down the machine",
			Action: WithClient(poweroff),
		},
	}
	var args []string
	command := path.Base(os.Args[0])
	if command == "corectl" {
		args = os.Args
	} else {
		args = []string{"corectl", command}
		args = append(args, os.Args[1:]...)
	}
	app.Run(args)
}
