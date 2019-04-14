package containers

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"

	"github.com/pborman/uuid"
	"github.com/threefoldtech/0-core/apps/core0/screen"
	"github.com/threefoldtech/0-core/apps/plugins/cgroup"
	"github.com/threefoldtech/0-core/apps/plugins/protocol"
	"github.com/threefoldtech/0-core/apps/plugins/socat"
	"github.com/threefoldtech/0-core/apps/plugins/zfs"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/settings"
	"github.com/threefoldtech/0-core/base/utils"
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

type ContainerDispatchArguments struct {
	Container uint16     `json:"container"`
	Command   pm.Command `json:"command"`
}

type containerPortForward struct {
	Container     uint16 `json:"container"`
	ContainerPort int    `json:"container_port"`
	HostPort      string `json:"host_port"`
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

	var ovs Container
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
	Container Container `json:"container"`
}

func (m *Manager) getByName(name string) *container {
	m.conM.RLock()
	defer m.conM.RUnlock()

	for _, c := range m.containers {
		if strings.EqualFold(c.Args.Name, name) {
			return c
		}
	}

	return nil
}

func (m *Manager) getByID(id uint16) *container {
	m.conM.RLock()
	defer m.conM.RUnlock()

	c, _ := m.containers[id]

	return c
}

func (m *Manager) get(ctx pm.Context) (interface{}, error) {
	var args struct {
		Query interface{} `json:"query"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}
	var cont *container

	switch query := args.Query.(type) {
	case string:
		cont = m.getByName(query)
	case float64:
		if query < 0 || query > math.MaxUint16 {
			return nil, pm.BadRequestError("query out of range")
		}

		cont = m.getByID(uint16(query))
	default:
		return nil, pm.BadRequestError("invalid query")
	}

	//not found
	if cont == nil {
		return nil, nil
	}

	ports, err := m.socat().List(cont.forwardId())
	if err != nil {
		return nil, err
	}

	cont.Args.Port = ports
	return cont, nil
}

func (m *Manager) list(ctx pm.Context) (interface{}, error) {
	containers := make(map[uint16]ContainerInfo)

	rules, err := m.socat().ListAll(socat.Container)
	if err != nil {
		return nil, err
	}

	m.conM.RLock()
	defer m.conM.RUnlock()
	for id, c := range m.containers {
		c.Args.Port, _ = rules[c.forwardId()]
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

func (m *Manager) GetWithTags(tags ...string) []Container {
	m.conM.RLock()
	defer m.conM.RUnlock()

	var result []Container
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

func (m *Manager) GetOneWithTags(tags ...string) Container {
	result := m.GetWithTags(tags...)
	if len(result) > 0 {
		return result[0]
	}

	return nil
}

func (m *Manager) Of(id uint16) Container {
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
