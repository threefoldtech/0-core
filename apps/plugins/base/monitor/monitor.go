package monitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	ps "github.com/shirou/gopsutil/process"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/utils"
	"github.com/vishvananda/netlink"
)

const (
	monitorDisk    = "disk"
	monitorCPU     = "cpu"
	monitorNetwork = "network"
	monitorMemory  = "memory"
)

var (
	networkMonitorTypes = []string{
		"device",
		"bridge",
		"openvswitch",
		"veth",
	}

	networkMonitorIgnoreNames = []string{
		"lo",
		"ovs-system",
	}
)

type Pair [2]string

type monitor struct{}

func (m *Manager) monitor(ctx pm.Context) (interface{}, error) {
	var args struct {
		Domain string `json:"domain"`
	}
	cmd := ctx.Command()

	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	switch strings.ToLower(args.Domain) {
	case monitorDisk:
		return nil, m.disk()
	case monitorCPU:
		return nil, m.cpu()
	case monitorMemory:
		return nil, m.memory()
	case monitorNetwork:
		return nil, m.network()
	default:
		return nil, fmt.Errorf("invalid monitoring domain: %s", args.Domain)
	}
}

func (m *Manager) disk() error {
	counters, err := disk.IOCounters()
	if err != nil {
		return err
	}

	for name, counter := range counters {
		m.api.Aggregate(pm.AggreagteDifference,
			"disk.iops.read",
			float64(counter.ReadCount),
			name, pm.Tag{"type", "phys"},
		)

		m.api.Aggregate(pm.AggreagteDifference,
			"disk.iops.write",
			float64(counter.WriteCount),
			name, pm.Tag{"type", "phys"},
		)

		m.api.Aggregate(pm.AggreagteDifference,
			"disk.throughput.read",
			float64(counter.ReadBytes/1024),
			name, pm.Tag{"type", "phys"},
		)

		m.api.Aggregate(pm.AggreagteDifference,
			"disk.throughput.write",
			float64(counter.WriteBytes/1024),
			name, pm.Tag{"type", "phys"},
		)
	}

	parts, err := disk.Partitions(false)
	if err != nil {
		return err
	}

	mounts := map[string]string{}
	//check the device only once, any mountpoint will do.
	for _, part := range parts {
		mounts[part.Device] = part.Mountpoint
	}

	for device, mount := range mounts {
		name := path.Base(device) //to be consistent with io counters.
		usage, err := disk.Usage(mount)
		if err != nil {
			log.Errorf("failed to get usage of '%s'", err)
			continue
		}

		m.api.Aggregate(pm.AggreagteAverage,
			"disk.size.total",
			float64(usage.Total),
			name,
			pm.Tag{"type", "phys"},
			pm.Tag{"fs", usage.Fstype},
		)

		m.api.Aggregate(pm.AggreagteAverage,
			"disk.size.free",
			float64(usage.Free),
			name,
			pm.Tag{"type", "phys"},
			pm.Tag{"fs", usage.Fstype},
		)
	}

	return nil
}

func (m *Manager) cpu() error {
	times, err := cpu.Times(true)
	if err != nil {
		return err
	}

	for nr, t := range times {
		m.api.Aggregate(pm.AggreagteDifference,
			"machine.CPU.utilisation",
			t.System+t.User,
			fmt.Sprint(nr), pm.Tag{"type", "phys"},
		)
	}

	percent, err := cpu.Percent(time.Second, true)
	if err != nil {
		return err
	}

	for nr, v := range percent {
		m.api.Aggregate(pm.AggreagteAverage,
			"machine.CPU.percent",
			v,
			fmt.Sprint(nr), pm.Tag{"type", "phys"},
		)
	}

	const StatFile = "/proc/stat"
	stat, err := ioutil.ReadFile(StatFile)
	if err != nil {
		return err
	}

	statmap := make(map[string]string)
	for _, line := range strings.Split(string(stat), "\n") {
		var key, value string
		if n, err := fmt.Sscanf(line, "%s %v", &key, &value); n == 2 && err == nil {
			statmap[key] = value
		}
	}

	if ctxt, ok := statmap["ctxt"]; ok {
		v, _ := strconv.ParseFloat(ctxt, 64)
		m.api.Aggregate(pm.AggreagteDifference,
			"machine.CPU.contextswitch",
			v,
			"", pm.Tag{"type", "phys"},
		)
	}

	if intr, ok := statmap["intr"]; ok {
		v, _ := strconv.ParseFloat(intr, 64)
		m.api.Aggregate(pm.AggreagteDifference,
			"machine.CPU.interrupts",
			v,
			"", pm.Tag{"type", "phys"},
		)
	}

	pids, _ := ps.Pids()
	var threads int32 = 0
	for _, pid := range pids {
		process, err := ps.NewProcess(pid)
		if err != nil {
			//probably process is gone
			continue
		}

		if num, err := process.NumThreads(); err == nil {
			threads += num
		}
	}

	m.api.Aggregate(pm.AggreagteAverage, "machine.process.threads", float64(threads), "", pm.Tag{"type", "phys"})

	return nil
}

func (m *Manager) memory() error {
	virt, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	m.api.Aggregate(pm.AggreagteAverage,
		"machine.memory.ram.available",
		float64(virt.Available)/(1024.*1024.),
		"", pm.Tag{"type", "phys"},
	)

	swap, err := mem.SwapMemory()
	if err != nil {
		return err
	}

	m.api.Aggregate(pm.AggreagteAverage,
		"machine.memory.swap.left",
		float64(swap.Free)/(1024.*1024.),
		"", pm.Tag{"type", "phys"},
	)

	m.api.Aggregate(pm.AggreagteAverage,
		"machine.memory.swap.used",
		float64(swap.Used)/(1024.*1024.),
		"", pm.Tag{"type", "phys"},
	)

	return nil
}

func (m *Manager) network() error {
	counters, err := net.IOCounters(true)
	if err != nil {
		return err
	}

	for _, counter := range counters {
		link, err := netlink.LinkByName(counter.Name)
		if err != nil {
			continue
		}

		if utils.InString(networkMonitorIgnoreNames, counter.Name) {
			continue
		}

		//only required devices
		if !utils.InString(networkMonitorTypes, link.Type()) ||
			(link.Type() == "veth" && !strings.HasPrefix(counter.Name, "contm")) {
			continue
		}

		m.api.Aggregate(pm.AggreagteDifference,
			"network.throughput.outgoing",
			float64(counter.BytesSent)/(1024.*1024.),
			counter.Name,
			pm.Tag{"type", "phys"}, pm.Tag{"kind", link.Type()},
		)

		m.api.Aggregate(pm.AggreagteDifference,
			"network.throughput.incoming",
			float64(counter.BytesRecv)/(1024.*1024.),
			counter.Name,
			pm.Tag{"type", "phys"}, pm.Tag{"kind", link.Type()},
		)

		m.api.Aggregate(pm.AggreagteDifference,
			"network.packets.tx",
			float64(counter.PacketsSent),
			counter.Name,
			pm.Tag{"type", "phys"}, pm.Tag{"kind", link.Type()},
		)

		m.api.Aggregate(pm.AggreagteDifference,
			"network.packets.rx",
			float64(counter.PacketsRecv),
			counter.Name,
			pm.Tag{"type", "phys"}, pm.Tag{"kind", link.Type()},
		)
	}

	return nil
}
