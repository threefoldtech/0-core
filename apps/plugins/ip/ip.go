package ip

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strings"
	"sync"

	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/vishvananda/netlink"
)

const (
	bondingBaseDir = "/proc/net/bonding/"
)

var (
	//Plugin entry point
	ip ipmgr

	Plugin = plugin.Plugin{
		Name:    "ip",
		Version: "1.0",
		Open: func(api plugin.API) error {
			ip.api = api
		},
		Actions: map[string]pm.Action{
			"bridge.add":   ip.brAdd,
			"bridge.del":   ip.brDel,
			"bridge.addif": ip.brAddInf,
			"bridge.delif": ip.brDelInf,

			"link.up":   ip.linkUp,
			"link.down": ip.linkDown,
			"link.name": ip.linkName,
			"link.link": ip.linkList,
			"link.mtu":  ip.linkMTU,

			"addr.add":  ip.addrAdd,
			"addr.del":  ip.addrDel,
			"addr.list": ip.addrList,

			"route.add":  ip.routeAdd,
			"route.del":  ip.routeDel,
			"route.list": ip.routeList,

			"bond.add":  ip.bondAdd,
			"bond.del":  ip.bondDel,
			"bond.list": ip.bondList,
		},
	}
)

type ipmgr struct {
	bondOnce sync.Once
	api      plugin.API
}

func (m *ipmgr) initBonding() {
	m.bondOnce.Do(func() {
		m.api.System("modprobe", "bonding")
		link, err := netlink.LinkByName("bond0")
		if err != nil {
			return
		}

		netlink.LinkDel(link)
	})
}

func (_ *ipmgr) parseBond(c string) interface{} {
	type M map[string]string
	type L []M

	m := make(M)
	l := make(L, 0)

	for _, line := range strings.Split(c, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			l = append(l, m)
			m = make(M)
			continue
		}

		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			continue
		}
		m[kv[0]] = kv[1]
	}

	if len(m) > 0 {
		l = append(l, m)
	}

	return l
}

func (m *ipmgr) bondList(ctx pm.Context) (interface{}, error) {
	m.initBonding()
	files, err := ioutil.ReadDir(bondingBaseDir)
	if err != nil {
		return nil, err
	}

	bonds := make(map[string]interface{})

	for _, info := range files {
		p := filepath.Join(bondingBaseDir, info.Name())
		bytes, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, pm.InternalError(err)
		}
		bonds[info.Name()] = m.parseBond(string(bytes))
	}

	return bonds, nil
}

func (m *ipmgr) bondDel(ctx pm.Context) (interface{}, error) {
	var args struct {
		Bond string `json:"bond"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(args.Bond)
	if err != nil {
		return nil, err
	}

	return nil, netlink.LinkDel(link)
}

func (m *ipmgr) bondAdd(ctx pm.Context) (interface{}, error) {
	m.initBonding()

	var args struct {
		Bond       string   `json:"bond"`
		Interfaces []string `json:"interfaces"`
		MTU        int      `json:"mtu"`
	}

	cmd := ctx.Command()

	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	mtu := args.MTU
	if mtu == 0 {
		mtu = 1500
	}

	bond := netlink.NewLinkBond(netlink.LinkAttrs{
		Name:   args.Bond,
		MTU:    mtu,
		TxQLen: 1000,
	})

	bond.Mode = netlink.BOND_MODE_BALANCE_RR
	bond.MTU = mtu

	enslave := []string{
		args.Bond,
	}

	for _, infName := range args.Interfaces {
		slave, err := netlink.LinkByName(infName)
		if err != nil {
			return nil, pm.NotFoundError(fmt.Errorf("interface %s: %s", infName, err))
		}
		if err := netlink.LinkSetMTU(slave, mtu); err != nil {
			return nil, err
		}
		if err := netlink.LinkSetUp(slave); err != nil {
			return nil, err
		}

		enslave = append(enslave, infName)
	}

	if err := netlink.LinkAdd(bond); err != nil {
		return nil, err
	}

	if err := netlink.LinkSetUp(bond); err != nil {
		return nil, err
	}

	if _, err := pm.System("ifenslave", enslave...); err != nil {
		return nil, err
	}

	return nil, nil
}

type LinkArguments struct {
	Name string `json:"name"`
}

type BridgeArguments struct {
	LinkArguments
	HwAddress string `json:"hwaddr"`
}

func (_ *ipmgr) brAdd(ctx pm.Context) (interface{}, error) {
	var args BridgeArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	var hw net.HardwareAddr

	if args.HwAddress != "" {
		var err error
		hw, err = net.ParseMAC(args.HwAddress)
		if err != nil {
			return nil, err
		}
	}

	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name:   args.Name,
			TxQLen: 1000,
		},
	}

	if err := netlink.LinkAdd(br); err != nil {
		return nil, err
	}

	var err error
	defer func() {
		if err != nil {
			netlink.LinkDel(br)
		}
	}()

	if args.HwAddress != "" {
		err = netlink.LinkSetHardwareAddr(br, hw)
	}

	return nil, err
}

func (_ *ipmgr) brDel(ctx pm.Context) (interface{}, error) {
	var args LinkArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(args.Name)
	if err != nil {
		return nil, err
	}
	if link.Type() != "bridge" {
		return nil, fmt.Errorf("no bridge with name '%s'", args.Name)
	}

	return nil, netlink.LinkDel(link)
}

type BridgeInfArguments struct {
	LinkArguments
	Inf string `json:"inf"`
}

func (_ *ipmgr) brAddInf(ctx pm.Context) (interface{}, error) {
	var args BridgeInfArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(args.Name)
	if err != nil {
		return nil, err
	}
	if link.Type() != "bridge" {
		return nil, fmt.Errorf("no bridge with name '%s'", args.Name)
	}

	inf, err := netlink.LinkByName(args.Inf)
	if err != nil {
		return nil, err
	}

	return nil, netlink.LinkSetMaster(inf, link.(*netlink.Bridge))
}

func (_ *ipmgr) brDelInf(ctx pm.Context) (interface{}, error) {
	var args BridgeInfArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(args.Name)
	if err != nil {
		return nil, err
	}
	if link.Type() != "bridge" {
		return nil, fmt.Errorf("no bridge with name '%s'", args.Name)
	}

	inf, err := netlink.LinkByName(args.Inf)
	if err != nil {
		return nil, err
	}

	if inf.Attrs().MasterIndex != link.Attrs().Index {
		return nil, fmt.Errorf("interface is not connected to bridge")
	}

	return nil, netlink.LinkSetNoMaster(inf)
}

func (_ *ipmgr) linkUp(ctx pm.Context) (interface{}, error) {
	var args LinkArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(args.Name)
	if err != nil {
		return nil, err
	}

	return nil, netlink.LinkSetUp(link)
}

type LinkNameArguments struct {
	LinkArguments
	New string `json:"new"`
}

type LinkMTUArguments struct {
	LinkArguments
	MTU int `json:"mtu"`
}

func (_ *ipmgr) linkName(ctx pm.Context) (interface{}, error) {
	var args LinkNameArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(args.Name)
	if err != nil {
		return nil, err
	}

	return nil, netlink.LinkSetName(link, args.New)
}

func (_ *ipmgr) linkMTU(ctx pm.Context) (interface{}, error) {
	var args LinkMTUArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, pm.BadRequestError(err)
	}

	link, err := netlink.LinkByName(args.Name)
	if err != nil {
		return nil, pm.NotFoundError(err)
	}

	return nil, netlink.LinkSetMTU(link, args.MTU)
}

func (_ *ipmgr) linkDown(ctx pm.Context) (interface{}, error) {
	var args LinkArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(args.Name)
	if err != nil {
		return nil, err
	}

	return nil, netlink.LinkSetDown(link)
}

type Link struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	HwAddr string `json:"hwaddr"`
	Master string `json:"master"`
	Up     bool   `json:"up"`
	MTU    int    `json:"mtu"`
}

func (_ *ipmgr) linkList(ctx pm.Context) (interface{}, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	result := make([]Link, 0)
	for _, link := range links {
		master := ""
		if link.Attrs().MasterIndex != 0 {
			for _, l := range links {
				if link.Attrs().MasterIndex == l.Attrs().Index {
					master = l.Attrs().Name
				}
			}
		}

		attrs := link.Attrs()

		result = append(result,
			Link{
				Type:   link.Type(),
				Name:   attrs.Name,
				HwAddr: attrs.HardwareAddr.String(),
				Master: master,
				Up:     attrs.Flags&net.FlagUp != 0,
				MTU:    attrs.MTU,
			},
		)
	}

	return result, nil
}

type AddrArguments struct {
	LinkArguments
	IP string `json:"ip"`
}

func (_ *ipmgr) addrAdd(ctx pm.Context) (interface{}, error) {
	var args AddrArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(args.Name)
	if err != nil {
		return nil, err
	}

	addr, err := netlink.ParseAddr(args.IP)
	if err != nil {
		return nil, err
	}

	return nil, netlink.AddrAdd(link, addr)
}

func (_ *ipmgr) addrDel(ctx pm.Context) (interface{}, error) {
	var args AddrArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(args.Name)
	if err != nil {
		return nil, err
	}

	addr, err := netlink.ParseAddr(args.IP)
	if err != nil {
		return nil, err
	}

	return nil, netlink.AddrDel(link, addr)
}

func (_ *ipmgr) addrList(ctx pm.Context) (interface{}, error) {
	var args LinkArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(args.Name)
	if err != nil {
		return nil, err
	}

	addr, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0)
	for _, addr := range addr {
		ips = append(ips, addr.IPNet.String())
	}

	return ips, err
}

type Route struct {
	Dev string `json:"dev"`
	Dst string `json:"dst"`
	Gw  string `json:"gw"`
}

func (r *Route) route() (*netlink.Route, error) {
	link, err := netlink.LinkByName(r.Dev)
	if err != nil {
		return nil, err
	}

	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
	}

	if r.Dst != "" {
		dst, err := netlink.ParseIPNet(r.Dst)
		if err != nil {
			return nil, err
		}
		route.Dst = dst
	}

	if r.Gw != "" {
		gw := net.ParseIP(r.Gw)
		if gw == nil {
			return nil, fmt.Errorf("invalid gw ip '%s'", r.Gw)
		}
		route.Gw = gw
	}

	return route, nil
}

func (_ *ipmgr) routeAdd(ctx pm.Context) (interface{}, error) {
	var args Route
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	route, err := args.route()
	if err != nil {
		return nil, err
	}

	return nil, netlink.RouteAdd(route)
}

func (_ *ipmgr) routeDel(ctx pm.Context) (interface{}, error) {
	var args Route
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	route, err := args.route()
	if err != nil {
		return nil, err
	}

	filter := netlink.RT_FILTER_OIF
	if route.Dst != nil {
		filter |= netlink.RT_FILTER_DST
	}
	if len(route.Gw) != 0 {
		filter |= netlink.RT_FILTER_GW
	}

	routes, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, route, filter)
	if err != nil {
		return nil, err
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("route not found")
	} else if len(routes) > 1 {
		return nil, fmt.Errorf("ambiguous route matches multiple routes")
	}

	return nil, netlink.RouteDel(&routes[0])
}

func (_ *ipmgr) routeList(ctx pm.Context) (interface{}, error) {
	routes, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		return nil, err
	}

	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	results := make([]Route, 0)
	for _, r := range routes {
		var dst, gw, dev string
		for _, l := range links {
			if r.LinkIndex == l.Attrs().Index {
				dev = l.Attrs().Name
				break
			}
		}

		if r.Dst != nil {
			dst = r.Dst.String()
		}
		if r.Gw != nil {
			gw = r.Gw.String()
		}

		results = append(results,
			Route{
				Dst: dst,
				Gw:  gw,
				Dev: dev,
			},
		)
	}

	return results, nil
}
