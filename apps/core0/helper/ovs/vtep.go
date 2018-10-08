package ovs

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

const (
	MultiCastGroup = 239
)

func getGroupForVNID(vnid uint) net.IP {
	//VNID is 24 bit, that fits the last 3 octet of the MC group IP
	id := (vnid / 256) + 1

	ip := fmt.Sprintf("%d.%d.%d.%d",
		MultiCastGroup,
		id&0x00ff0000>>16,
		id&0x0000ff00>>8,
		id&0x000000ff,
	)

	return net.ParseIP(ip)
}

//VtepEnsure ensures a vtep with given vnid and master (bridge)
func VtepEnsure(vnid uint, bridge string) (string, error) {
	dev, err := netlink.LinkByName(bridge)

	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("vtep%d", vnid)
	link, err := netlink.LinkByName(name)

	if err == nil {
		if link.Type() != "vxlan" {
			return name, fmt.Errorf("invalid device type got '%s'", link.Type())
		}

		if link.(*netlink.Vxlan).VtepDevIndex != dev.Attrs().Index {
			return name, fmt.Errorf("reassigning vxlan to another master bridge is not allowed")
		}

		return name, nil
	}

	vxlan := &netlink.Vxlan{
		LinkAttrs: netlink.LinkAttrs{
			Name:   name,
			Flags:  net.FlagBroadcast | net.FlagMulticast,
			MTU:    1500,
			TxQLen: -1,
		},
		VxlanId:      int(vnid),
		Group:        getGroupForVNID(vnid),
		VtepDevIndex: dev.Attrs().Index,
		Learning:     true,
	}

	if err := netlink.LinkAdd(vxlan); err != nil {
		return name, err
	}

	return name, netlink.LinkSetUp(vxlan)
}

//VtepDelete delets a vtep with vnid
func VtepDelete(vnid uint) error {
	name := fmt.Sprintf("vtep%d", vnid)
	link, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}

	return netlink.LinkDel(link)
}
