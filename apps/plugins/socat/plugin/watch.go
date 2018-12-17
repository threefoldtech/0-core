package main

import (
	"strings"

	"github.com/threefoldtech/0-core/apps/plugins/nft"
	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/stream"
	"github.com/threefoldtech/0-core/base/utils"
	"github.com/vishvananda/netlink"
)

func getCurrentIPs() ([]string, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, err
	}

	var all []string
	for _, link := range links {
		name := link.Attrs().Name
		if name == "lo" || !utils.InString([]string{"device", "tun", "tap", "openvswitch", "bridge", "bond"}, link.Type()) {
			continue
		}

		if strings.HasPrefix(name, "tun") || strings.HasPrefix(name, "tap") {
			continue
		}

		ips, _ := netlink.AddrList(link, netlink.FAMILY_V4)
		for _, ip := range ips {
			all = append(all, ip.IP.String())
		}
	}

	return all, nil
}

type notifyHook struct {
	pm.NOOPHook
	Notify func()
}

func (n *notifyHook) Message(msg *stream.Message) {
	if n != nil {
		n.Notify()
	}
}

func (s *socatManager) monitorIPChangesUpdateSocat() error {
	var current map[string]struct{}

	notify := func() {
		log.Debugf("updating dnat host ips")

		ips, err := getCurrentIPs()
		if err != nil {
			log.Errorf("failed to get active ips: %s", err)
			return
		}

		for _, ip := range ips {
			if _, ok := current[ip]; ok {
				//ip already configure
				delete(current, ip)
			}
		}

		for ip := range current {
			if err := s.nft.IPv4SetDel(nft.FamilyIP, "nat", "host", ip); err != nil {
				log.Errorf("failed to delete host ip(%s): %s", ip, err)
			}
		}

		current = make(map[string]struct{})

		for _, ip := range ips {
			current[ip] = struct{}{}
		}

		if err := s.nft.IPv4Set(nft.FamilyIP, "nat", "host", ips...); err != nil {
			log.Errorf("failed to set host ips(%s): %s", strings.Join(ips, ","), err)
		}
	}

	_, err := s.api.Run(&pm.Command{
		ID:      "socat.notify",
		Command: pm.CommandSystem,
		Flags: pm.JobFlags{
			Protected: true,
		},
		Arguments: pm.MustArguments(
			pm.SystemCommandArguments{
				Name: "ip",
				Args: []string{"monitor", "address"},
			},
		),
	}, &notifyHook{
		Notify: notify,
	})

	return err
}
