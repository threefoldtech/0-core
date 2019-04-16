package containers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"syscall"

	"github.com/threefoldtech/0-core/base/pm"
	"github.com/vishvananda/netlink"
)

type ContainerCreateArguments struct {
	Root        string            `json:"root"`         //Root plist
	Mount       map[string]string `json:"mount"`        //data disk mounts.
	HostNetwork bool              `json:"host_network"` //share host networking stack
	Identity    string            `json:"identity"`     //zerotier identity
	Nics        []*Nic            `json:"nics"`         //network setup (only respected if HostNetwork is false)
	Port        map[string]int    `json:"port"`         //port forwards (only if default networking is enabled)
	Privileged  bool              `json:"privileged"`   //Apply cgroups and capabilities limitations on the container
	Hostname    string            `json:"hostname"`     //hostname
	Storage     string            `json:"storage"`      //ardb storage needed for g8ufs mounts.
	Name        string            `json:"name"`         //for searching containers
	Tags        pm.Tags           `json:"tags"`         //for searching containers
	Env         map[string]string `json:"env"`          //environment variables.
	CGroups     []CGroup          `json:"cgroups"`      //container creation cgroups
	Config      map[string]string `json:"config"`       //overrides container config (from flist)
}

func (c *ContainerCreateArguments) Validate(m *Manager) error {
	if c.Root == "" {
		return fmt.Errorf("root plist is required")
	}

	for host, guest := range c.Mount {
		u, err := url.Parse(host)
		if err != nil {
			return fmt.Errorf("invalid host mount: %s", err)
		}
		if u.Scheme != "" {
			//probably a plist
			continue
		}
		p := u.Path
		if !path.IsAbs(p) {
			return fmt.Errorf("host path '%s' must be absolute", host)
		}
		if !path.IsAbs(guest) {
			return fmt.Errorf("guest path '%s' must be absolute", guest)
		}
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return fmt.Errorf("host path '%s' does not exist", p)
		}
	}

	for host, guest := range c.Port {
		if !m.socat().ValidHost(host) {
			return fmt.Errorf("invalid host port '%s'", host)
		}
		if guest < 0 || guest > 65535 {
			return fmt.Errorf("invalid guest port '%d'", guest)
		}
	}

	//validating networking
	brcounter := make(map[string]int)
	for _, nic := range c.Nics {
		if nic.State == NicStateDestroyed {
			continue
		}
		switch nic.Type {
		case "default":
			brcounter[DefaultBridgeName]++
			if brcounter[DefaultBridgeName] > 1 {
				return fmt.Errorf("only one default network is allowed")
			}
		case "bridge":
			if nic.ID == DefaultBridgeName {
				return fmt.Errorf("cannot use bridge %s with nic type 'bridge', please use type default instead", DefaultBridgeName)
			}
			brcounter[nic.ID]++
			if brcounter[nic.ID] > 1 {
				return fmt.Errorf("connecting to bridge '%s' more than one time is not allowed", nic.ID)
			}
		case "passthrough":
			fallthrough
		case "macvlan":
			l, err := netlink.LinkByName(nic.ID)
			if err != nil {
				return err
			}
			ltype := l.Type()

			if ltype != "device" && ltype != "dummy" {
				return fmt.Errorf("cannot use %s %s with nic type '%s', please use link with type 'device' instead", ltype, nic.ID, nic.Type)
			}
			brcounter[nic.ID]++
			if brcounter[nic.ID] > 1 {
				return fmt.Errorf("connecting to link '%s' more than one time is not allowed", nic.ID)
			}
		case "vlan":
		case "vxlan":
		case "zerotier":
		default:
			return fmt.Errorf("unsupported network type '%s'", nic.Type)
		}
	}

	nameset := make(map[string]byte)
	for _, nic := range c.Nics {
		if nic.State == NicStateDestroyed {
			continue
		}
		if nic.Name != "" {
			if _, ok := nameset[nic.Name]; ok {
				return fmt.Errorf("name '%v' is passed twice in the container", nic.Name)
			} else {
				nameset[nic.Name] = 1
			}
			if len(nic.Name) > 15 { //linux limit on interface name
				return fmt.Errorf("invalid name '%s' too long", nic.Name)
			}
			if nic.Name == "default" { //probably we need to expand this list with more reserved names
				//`default` is not allowed by linux for some reason.
				return fmt.Errorf("invalid name `%s`", nic.Name)
			}
			//avoid conflict with eth or zt
			if strings.HasPrefix(nic.Name, "eth") || strings.HasPrefix(nic.Name, "zt") {
				return fmt.Errorf("name '%v' cannot be used as it is started with eth or zt", nic.Name)
			}
		}
	}

	for _, cgroup := range c.CGroups {
		if !m.cgroup().Exists(cgroup.Subsystem(), cgroup.Name()) {
			return fmt.Errorf("invalid cgroup %v", cgroup)
		}
	}

	return nil
}

//ContainerConfig manages container configuration (state) on disk
type ContainerConfig struct {
	ContainerCreateArguments
	file *os.File
}

//LoadConfig from file
func LoadConfig(path string) (*ContainerConfig, error) {
	config, err := newConfig(path)
	if err != nil {
		return nil, err
	}

	if err := config.load(); err != nil {
		config.Release()
		return nil, err
	}

	return config, nil
}

//NewConfig creates a new config file that has args as initial content
func NewConfig(path string, args ContainerCreateArguments) (*ContainerConfig, error) {
	config, err := newConfig(path)
	if err != nil {
		return nil, err
	}

	config.ContainerCreateArguments = args
	if err := config.Write(); err != nil {
		config.Release()
		return nil, err
	}

	return config, nil

}

func newConfig(path string) (*ContainerConfig, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		file.Close()
		return nil, err
	}

	config := ContainerConfig{
		file: file,
	}

	return &config, nil
}

func (a *ContainerConfig) Write() error {
	if err := a.file.Truncate(0); err != nil {
		return err
	}
	enc := json.NewEncoder(a.file)
	enc.SetIndent("", " ")

	if err := enc.Encode(a.ContainerCreateArguments); err != nil {
		return err
	}

	return a.file.Sync()
}

//WriteRelease write and release. the release is performed even if the write fails
func (a *ContainerConfig) WriteRelease() error {
	defer a.Release()
	return a.Write()
}
func (a *ContainerConfig) load() error {
	if _, err := a.file.Seek(0, 0); err != nil {
		return err
	}

	dec := json.NewDecoder(a.file)
	return dec.Decode(&a.ContainerCreateArguments)
}

//Release releases a container config file
func (a *ContainerConfig) Release() error {
	syscall.Flock(int(a.file.Fd()), syscall.LOCK_UN)
	return a.file.Close()
}
