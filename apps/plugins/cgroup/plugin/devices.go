package main

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/threefoldtech/0-core/apps/plugins/cgroup"
)

type DevicesGroup interface {
	cgroup.Group
	Deny(spec string) error
	Allow(spec string) error
	List() ([]string, error)
}

func mkDevicesGroup(name string, subsys cgroup.Subsystem) cgroup.Group {
	return &devicesCGroup{
		Group{name: name, subsys: subsys},
	}
}

type devicesCGroup struct {
	Group
}

func (g *devicesCGroup) Deny(spec string) error {
	p := path.Join(g.base(), "devices.deny")
	return ioutil.WriteFile(p, []byte(spec), 0200)
}

func (g *devicesCGroup) Allow(spec string) error {
	p := path.Join(g.base(), "devices.allow")
	return ioutil.WriteFile(p, []byte(spec), 0200)
}

func (g *devicesCGroup) List() ([]string, error) {
	p := path.Join(g.base(), "devices.list")
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(data), "\n"), nil
}

func (g *devicesCGroup) Root() cgroup.Group {
	return &devicesCGroup{
		Group: Group{subsys: g.subsys},
	}
}

func (g *devicesCGroup) Reset() {

}

var _ DevicesGroup = &devicesCGroup{}
