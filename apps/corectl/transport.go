package main

import (
	"fmt"
	"strconv"

	"github.com/codegangsta/cli"
	client "github.com/threefoldtech/0-core/client/go-client"
)

func getContainerClient(cl client.Client, idOrTag string) (client.Client, error) {
	mgr := client.Container(cl)
	if id, err := strconv.ParseInt(idOrTag, 10, 64); err == nil {
		//valid id
		return mgr.Client(int(id)), nil
	}

	//else, assume a tag
	results, err := mgr.List(idOrTag)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no container found with given tag")
	} else if len(results) != 1 {
		return nil, fmt.Errorf("tag matches multiple containers, please refine or use id")
	}

	var id int16
	for id = range results {
		//take the only available key value
	}

	return mgr.Client(int(id)), nil
}

func WithClient(action func(cl client.Client, c *cli.Context)) cli.ActionFunc {
	return func(c *cli.Context) error {
		cl, err := client.NewClient(c.GlobalString("socket"), "")
		if err != nil {
			log.Fatal(err)
		}

		if idOrTag := c.GlobalString("container"); len(idOrTag) != 0 {
			cl, err = getContainerClient(cl, idOrTag)
			if err != nil {
				log.Fatal(err)
			}
		}

		action(cl, c)
		return nil
	}
}
