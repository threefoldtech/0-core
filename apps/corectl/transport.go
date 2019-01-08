package main

import (
	"fmt"
	"strconv"

	"github.com/codegangsta/cli"
	client "github.com/threefoldtech/0-core/client/go-client"
)

// type Command struct {
// 	Sync      bool       `json:"sync"`
// 	Container string     `json:"container"`
// 	Content   pm.Command `json:"content"`
// }

// type TransportOptions struct {
// 	Timeout int
// 	ID      string
// }

// type Transport interface {
// 	Run(cmd Command) (*Response, error)
// }

// type unixSocketTransport struct {
// 	con *net.UnixConn
// 	opt *TransportOptions
// 	m   sync.Mutex
// }

// func NewUnixSocketTransport(n string, opt *TransportOptions) (Transport, error) {
// 	addr, err := net.ResolveUnixAddr("unix", n)
// 	if err != nil {
// 		return nil, err
// 	}

// 	con, err := net.DialUnix("unix", nil, addr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &unixSocketTransport{
// 		con: con,
// 		opt: opt,
// 	}, nil
// }

// func NewTransport(c *cli.Context) (Transport, error) {
// 	return NewUnixSocketTransport(c.GlobalString("socket"), &TransportOptions{
// 		Timeout: c.GlobalInt("timeout"),
// 		ID:      c.GlobalString("id"),
// 	})
// }

// func (t *unixSocketTransport) setDefaults(cmd *Command) error {
// 	if t.opt == nil {
// 		return nil
// 	}

// 	cmd.Content.MaxTime = t.opt.Timeout
// 	return nil
// }

// func (t *unixSocketTransport) Run(cmd Command) (*Response, error) {
// 	t.m.Lock()
// 	defer t.m.Unlock()

// 	if t.opt.ID == "" {
// 		cmd.Content.ID = uuid.New()
// 	} else {
// 		cmd.Content.ID = t.opt.ID
// 	}

// 	if err := t.setDefaults(&cmd); err != nil {
// 		return nil, err
// 	}

// 	data, err := json.Marshal(cmd)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if _, err := t.con.Write(data); err != nil {
// 		return nil, err
// 	}

// 	decoder := json.NewDecoder(t.con)
// 	var response Response
// 	if err := decoder.Decode(&response); err != nil {
// 		return nil, err
// 	}

// 	response.ID = cmd.Content.ID
// 	return &response, nil
// }

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
