package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/pborman/uuid"
	"github.com/threefoldtech/0-core/apps/core0/screen"
	"github.com/threefoldtech/0-core/apps/plugins/cgroup"
	"github.com/threefoldtech/0-core/apps/plugins/containers"
	"github.com/threefoldtech/0-core/apps/plugins/protocol"
	"github.com/threefoldtech/0-core/apps/plugins/socat"
	"github.com/threefoldtech/0-core/apps/plugins/zfs"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/settings"
	"github.com/threefoldtech/0-core/base/utils"
	"github.com/vishvananda/netlink"
)

const (
	coreXResponseQueue = "corex:results"
	coreXBinaryName    = "coreX"

	redisSocketSrc      = "/var/run/redis.socket"
	DefaultBridgeName   = "core0"
	ContainersHardLimit = 1000
)

const (
	NicStateConfigured = NicState("configured")
	NicStateDestroyed  = NicState("destroyed")
	NicStateUnknown    = NicState("unknown")
	NicStateError      = NicState("error")
)

var (
	BridgeIP          = []byte{172, 18, 0, 1}
	DefaultBridgeIP   = fmt.Sprintf("%d.%d.%d.%d", BridgeIP[0], BridgeIP[1], BridgeIP[2], BridgeIP[3])
	DefaultBridgeCIDR = fmt.Sprintf("%s/16", DefaultBridgeIP)
	DevicesCGroup     = CGroup{string(cgroup.DevicesSubsystem), "corex"}
)

type NetworkConfig struct {
	Dhcp    bool     `json:"dhcp"`
	CIDR    string   `json:"cidr"`
	Gateway string   `json:"gateway"`
	DNS     []string `json:"dns"`
}

type NicState string

type Nic struct {
	Type      string        `json:"type"`
	ID        string        `json:"id"`
	HWAddress string        `json:"hwaddr"`
	Name      string        `json:"name,omitempty"`
	Config    NetworkConfig `json:"config"`
	Monitor   bool          `json:"monitor"`
	State     NicState      `json:"state"`

	Index             int              `json:"-"`
	OriginalHWAddress net.HardwareAddr `json:"-"`
}

//CGroup defition
type CGroup [2]string

func (c CGroup) Subsystem() cgroup.Subsystem {
	return cgroup.Subsystem(c[0])
}

func (c CGroup) Name() string {
	return c[1]
}

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

type ContainerDispatchArguments struct {
	Container uint16     `json:"container"`
	Command   pm.Command `json:"command"`
}

type containerPortForward struct {
	Container     uint16 `json:"container"`
	ContainerPort int    `json:"container_port"`
	HostPort      string `json:"host_port"`
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

type Manager struct {
	api plugin.API

	sequence uint16
	seqM     sync.Mutex

	containers map[uint16]*container
	conM       sync.RWMutex
}

func (m *Manager) cgroup() cgroup.API {
	return m.api.MustPlugin("cgroup").(cgroup.API)
}

func (m *Manager) socat() socat.API {
	return m.api.MustPlugin("socat").(socat.API)
}

func (m *Manager) filesystem() zfs.API {
	return m.api.MustPlugin("zfs").(zfs.API)
}

func (m *Manager) protocol() protocol.API {
	return m.api.MustPlugin("protocol").(protocol.API)
}

func (m *Manager) logger() pm.MessageHandler {
	return m.api.MustPlugin("logger").(pm.MessageHandler)
}

/*
WARNING:
	Code here assumes that redis-server is started by core0 by the configuration files. If it wasn't started or failed
	to start, commands like core.create, core.dispatch, etc... will fail.
TODO:
	May be make redis-server start part of the bootstrap process without the need to depend on external configuration
	to run it.
*/

func (m *Manager) setUpCGroups() error {
	devices, err := m.cgroup().GetGroup(DevicesCGroup.Subsystem(), DevicesCGroup.Name())
	if err != nil {
		return err
	}

	if devices, ok := devices.(cgroup.DevicesGroup); ok {
		devices.Deny("a")
		for _, spec := range []string{
			"c 1:5 rwm",
			"c 1:3 rwm",
			"c 1:9 rwm",
			"c 1:7 rwm",
			"c 1:8 rwm",
			"c 5:0 rwm",
			"c 5:1 rwm",
			"c 5:2 rwm",
			"c *:* m",
			"c 136:* rwm",
			"c 10:200 rwm",
		} {
			devices.Allow(spec)
		}
	} else {
		return fmt.Errorf("failed to setup devices cgroups")
	}

	return nil
}

func (m *Manager) setUpDefaultBridge() error {
	return m.api.Internal("bridge.create", pm.M{
		"name": DefaultBridgeName,
		"network": pm.M{
			"nat":  true,
			"mode": "static",
			"settings": pm.M{
				"cidr": DefaultBridgeCIDR,
			},
		},
	}, nil)
}

func (m *Manager) getNextSequence() uint16 {
	m.seqM.Lock()
	defer m.seqM.Unlock()
	//get a read lock on the container dict as well
	m.conM.RLock()
	defer m.conM.RUnlock()

	for {
		m.sequence += 1
		if m.sequence != 0 && m.sequence < math.MaxUint16 {
			if _, ok := m.containers[m.sequence]; !ok {
				break
			}
		}
	}
	return m.sequence
}

func (m *Manager) setContainer(id uint16, c *container) {
	m.conM.Lock()
	defer m.conM.Unlock()
	m.containers[id] = c
	screen.Refresh()
}

//cleanup is called when a container terminates.
func (m *Manager) unsetContainer(id uint16) {
	m.conM.Lock()
	defer m.conM.Unlock()
	delete(m.containers, id)
	screen.Refresh()
}

func (m *Manager) nicAdd(ctx pm.Context) (interface{}, error) {
	var args struct {
		Container uint16 `json:"container"`
		Nic       Nic    `json:"nic"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	m.conM.RLock()
	defer m.conM.RUnlock()
	container, ok := m.containers[args.Container]
	if !ok {
		return nil, pm.NotFoundError(fmt.Errorf("container does not exist"))
	}

	if container.Args.HostNetwork {
		return nil, pm.BadRequestError(fmt.Errorf("cannot add a nic in host network mode"))
	}

	args.Nic.State = NicStateUnknown

	idx := len(container.Args.Nics)
	container.Args.Nics = append(container.Args.Nics, &args.Nic)

	if err := container.Args.Validate(m); err != nil {
		l := container.Args.Nics
		container.Args.Nics = l[:len(l)-1]
		return nil, pm.BadRequestError(err)
	}

	if err := container.preStartNetwork(idx, &args.Nic); err != nil {
		return nil, err
	}

	if err := container.postStartNetwork(idx, &args.Nic); err != nil {
		return nil, err
	}

	return nil, nil
}

func (m *Manager) nicRemove(ctx pm.Context) (interface{}, error) {
	var args struct {
		Container uint16 `json:"container"`
		Index     int    `json:"index"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	m.conM.RLock()
	defer m.conM.RUnlock()
	container, ok := m.containers[args.Container]
	if !ok {
		return nil, pm.NotFoundError(fmt.Errorf("container does not exist"))
	}

	if args.Index < 0 || args.Index >= len(container.Args.Nics) {
		return nil, pm.BadRequestError(fmt.Errorf("nic index out of range"))
	}
	nic := container.Args.Nics[args.Index]
	if nic.State != NicStateConfigured {
		return nil, pm.PreconditionFailedError(fmt.Errorf("nic is in '%s' state", nic.State))
	}

	if nic.Type == "zerotier" {
		//special handling for zerotier networks
		if err := container.leaveZerotierNetwork(args.Index, nic.ID); err != nil {
			nic.State = NicStateError
			return nil, err
		}

		nic.State = NicStateDestroyed
		return nil, nil
	}

	if nic.Type == "macvlan" {
		return nil, container.unLink(args.Index, nic)
	}

	var ovs containers.Container
	if nic.Type == "vlan" || nic.Type == "vxlan" {
		ovs = m.GetOneWithTags("ovs")
	}

	return nil, container.unBridge(args.Index, nic, ovs)
}

func (m *Manager) createContainer(args ContainerCreateArguments) (*container, error) {
	if err := args.Validate(m); err != nil {
		return nil, err
	}

	m.conM.RLock()
	count := len(m.containers)
	m.conM.RUnlock()
	limit := settings.Settings.Containers.MaxCount
	if limit == 0 {
		limit = ContainersHardLimit
	}

	if count >= limit {
		return nil, pm.ServiceUnavailableError(fmt.Errorf("reached the hard limit of %d containers", count))
	}

	id := m.getNextSequence()
	c := newContainer(m, id, args)
	m.setContainer(id, c)

	if _, err := c.Start(); err != nil {
		return nil, err
	}

	return c, nil
}

func (m *Manager) createSync(ctx pm.Context) (interface{}, error) {
	var args ContainerCreateArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		log.Errorf("invalid container params: %s", err)
		return nil, err
	}

	args.Tags = cmd.Tags
	container, err := m.createContainer(args)
	if err != nil {
		log.Errorf("failed to start container: %s", err)
		return nil, err
	}

	//after waiting we probably need to return the full result!
	return container.runner.Wait(), nil
}

func (m *Manager) create(ctx pm.Context) (interface{}, error) {
	var args ContainerCreateArguments
	cmd := ctx.Command()

	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	args.Tags = cmd.Tags
	container, err := m.createContainer(args)
	if err != nil {
		return nil, err
	}

	return container.id, nil
}

type ContainerInfo struct {
	pm.ProcessStats
	Container containers.Container `json:"container"`
}

func (m *Manager) list(ctx pm.Context) (interface{}, error) {
	containers := make(map[uint16]ContainerInfo)

	m.conM.RLock()
	defer m.conM.RUnlock()
	for id, c := range m.containers {
		name := fmt.Sprintf("core-%d", id)
		job, ok := m.api.JobOf(name)
		if !ok {
			continue
		}
		ps := job.Process()
		var state pm.ProcessStats
		if ps != nil {
			if stater, ok := ps.(pm.Stater); ok {
				state = *(stater.Stats())
			}
		}
		containers[id] = ContainerInfo{
			ProcessStats: state,
			Container:    c,
		}
	}

	return containers, nil
}

func (m *Manager) getCoreXQueue(id uint16) string {
	return fmt.Sprintf("core:%v", id)
}

func (m *Manager) pushToContainer(container *container, cmd *pm.Command) error {
	m.protocol().Flag(cmd.ID)
	return container.dispatch(cmd)
}

func (m *Manager) dispatch(ctx pm.Context) (interface{}, error) {
	var args ContainerDispatchArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	if args.Container <= 0 {
		return nil, fmt.Errorf("invalid container id")
	}

	m.conM.RLock()
	cont, ok := m.containers[args.Container]
	m.conM.RUnlock()

	if !ok {
		return nil, fmt.Errorf("container does not exist")
	}

	if args.Command.ID == "" {
		args.Command.ID = uuid.New()
	}

	if err := m.pushToContainer(cont, &args.Command); err != nil {
		return nil, err
	}

	return args.Command.ID, nil
}

//Dispatch command to container with ID (id)
func (m *Manager) Dispatch(id uint16, cmd *pm.Command) (*pm.JobResult, error) {
	cmd.ID = uuid.New()

	m.conM.RLock()
	cont, ok := m.containers[id]
	m.conM.RUnlock()

	if !ok {
		return nil, fmt.Errorf("container does not exist")
	}

	if err := m.pushToContainer(cont, cmd); err != nil {
		return nil, err
	}

	return m.protocol().Get(cmd.ID, 300)
}

type ContainerArguments struct {
	Container uint16 `json:"container"`
}

func (m *Manager) terminate(ctx pm.Context) (interface{}, error) {
	var args ContainerArguments
	cmd := ctx.Command()

	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}
	m.conM.RLock()
	container, ok := m.containers[args.Container]
	m.conM.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no container with id '%d'", args.Container)
	}

	return nil, container.Terminate()
}

type ContainerFindArguments struct {
	Tags []string `json:"tags"`
}

func (m *Manager) find(ctx pm.Context) (interface{}, error) {
	var args ContainerFindArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	containers := m.GetWithTags(args.Tags...)
	result := make(map[uint16]ContainerInfo)
	for _, c := range containers {
		name := fmt.Sprintf("core-%d", c.ID())
		job, ok := m.api.JobOf(name)
		if !ok {
			continue
		}
		ps := job.Process()
		var state pm.ProcessStats
		if ps != nil {
			if stater, ok := ps.(pm.Stater); ok {
				state = *(stater.Stats())
			}
		}

		result[c.ID()] = ContainerInfo{
			ProcessStats: state,
			Container:    c,
		}
	}

	return result, nil
}

func (m *Manager) GetWithTags(tags ...string) []containers.Container {
	m.conM.RLock()
	defer m.conM.RUnlock()

	var result []containers.Container
loop:
	for _, c := range m.containers {
		for _, tag := range tags {
			if !utils.InString(c.Args.Tags, tag) {
				continue loop
			}
		}
		result = append(result, c)
	}

	return result
}

func (m *Manager) GetOneWithTags(tags ...string) containers.Container {
	result := m.GetWithTags(tags...)
	if len(result) > 0 {
		return result[0]
	}

	return nil
}

func (m *Manager) Of(id uint16) containers.Container {
	m.conM.RLock()
	defer m.conM.RUnlock()
	cont, _ := m.containers[id]
	return cont
}

func (m *Manager) portforwardAdd(ctx pm.Context) (interface{}, error) {
	var args containerPortForward
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	m.conM.RLock()
	defer m.conM.RUnlock()

	container, ok := m.containers[args.Container]
	if !ok {
		return nil, pm.NotFoundError(fmt.Errorf("container does not exist"))
	}
	var defaultNic bool
	for _, nic := range container.Args.Nics {
		if nic.Type == "default" {
			defaultNic = true
			break
		}
	}
	if !defaultNic {
		return nil, fmt.Errorf("Container doesn't have a default nic")
	}

	if err := container.setPortForward(args.HostPort, args.ContainerPort); err != nil {
		return nil, err
	}

	return nil, nil
}

func (m *Manager) portforwardRemove(ctx pm.Context) (interface{}, error) {
	var args containerPortForward
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	m.conM.RLock()
	defer m.conM.RUnlock()

	container, ok := m.containers[args.Container]
	if !ok {
		return nil, pm.NotFoundError(fmt.Errorf("container does not exist"))
	}

	if err := m.socat().RemovePortForward(container.forwardId(), args.HostPort, args.ContainerPort); err != nil {
		return nil, err
	}

	return nil, nil
}

func (m *Manager) flistLayer(ctx pm.Context) (interface{}, error) {
	var args struct {
		Container uint16 `json:"container"`
		FList     string `json:"flist"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	m.conM.RLock()
	defer m.conM.RUnlock()

	container, ok := m.containers[args.Container]
	if !ok {
		return nil, pm.NotFoundError(fmt.Errorf("container does not exist"))
	}

	return nil, container.mergeFList(args.FList)
}
