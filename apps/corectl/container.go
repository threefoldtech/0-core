package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/olekukonko/tablewriter"
	client "github.com/threefoldtech/0-core/client/go-client"
	"gopkg.in/yaml.v2"
)

type containerData struct {
	Container struct {
		Arguments struct {
			Root     string   `json:"root"`
			Hostname string   `json:"hostname"`
			Tags     []string `json:"tags"`
		} `json:"arguments"`
		PID  int    `json:"pid"`
		Root string `json:"root"`
	} `json:"container"`
}

func containers(t client.Client, c *cli.Context) {
	var tags []string
	if c.Args().Present() {
		tags = append(tags, c.Args().First())
		tags = append(tags, c.Args().Tail()...)
	}

	cont := client.Container(t)

	containers, err := cont.List(tags...)

	if err != nil {
		log.Fatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"ID", "FLIST", "HOSTNAME", "TAGS"})
	ids := make([]int, 0, len(containers))
	for id := range containers {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)

	for _, id := range ids {
		container := containers[int16(id)]
		table.Append([]string{
			fmt.Sprint(id),
			container.Container.Arguments.Root,
			container.Container.Arguments.Hostname,
			strings.Join(container.Container.Arguments.Tags, ", "),
		})
	}

	table.Render()
}

func containerInspect(t client.Client, c *cli.Context) {
	idstr := c.Args().First()
	if idstr == "" {
		log.Fatal("missing container id")
	}
	id, err := strconv.ParseInt(idstr, 10, 16)
	if err != nil {
		log.Fatal(err)
	}

	cont := client.Container(t)

	containers, err := cont.List()

	if err != nil {
		log.Fatal(err)
	}

	container, ok := containers[int16(id)]
	if !ok {
		log.Fatalf("no container with id: %s", id)
	}

	data, _ := yaml.Marshal(container)
	fmt.Println(string(data))
}
