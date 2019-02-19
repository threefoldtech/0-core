package cgroup

import (
	"io/ioutil"
	"path"
	"strings"
)

func mkDevicesGroup(name string, subsys Subsystem) Group {
	return &devicesCGroup{
		cGroup{name: name, subsys: subsys},
	}
}

type devicesCGroup struct {
	cGroup
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

func (g *devicesCGroup) Root() Group {
	return &devicesCGroup{
		cGroup: cGroup{subsys: g.subsys},
	}
}

func (g *devicesCGroup) Reset() {

}

var _ DevicesGroup = &devicesCGroup{}
