package cgroup

import (
	"io/ioutil"
	"path"
	"strings"
)

func mkCPUSetGroup(name string, subsys Subsystem) Group {
	return &cpusetCGroup{
		cGroup{name: name, subsys: subsys},
	}
}

type cpusetCGroup struct {
	cGroup
}

//reset copies the default values from the root group. It sounds like
//this should be handled by the linux kernel, but it does not happen
//for the cpuset subsystem
//TODO: should we call this on the group creation ?
func (c *cpusetCGroup) Reset() {
	root := c.Root().(CPUSetGroup)

	spec, _ := root.GetCpus()
	c.Cpus(spec)

	spec, _ = root.GetMems()
	c.Mems(spec)
}

func (c *cpusetCGroup) Cpus(spec string) error {
	log.Debugf("setting cpu specs to: '%s'", spec)
	return ioutil.WriteFile(path.Join(c.base(), "cpuset.cpus"), []byte(spec), 0644)
}

func (c *cpusetCGroup) Mems(spec string) error {
	return ioutil.WriteFile(path.Join(c.base(), "cpuset.mems"), []byte(spec), 0644)
}

func (c *cpusetCGroup) GetCpus() (string, error) {
	data, err := ioutil.ReadFile(path.Join(c.base(), "cpuset.cpus"))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func (c *cpusetCGroup) GetMems() (string, error) {
	data, err := ioutil.ReadFile(path.Join(c.base(), "cpuset.mems"))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

func (c *cpusetCGroup) Root() Group {
	return &cpusetCGroup{
		cGroup: cGroup{subsys: c.subsys},
	}
}

var _ CPUSetGroup = &cpusetCGroup{}
