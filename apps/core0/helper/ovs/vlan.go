package ovs

import (
	"fmt"
)

func in(l []string, a string) bool {
	for _, i := range l {
		if i == a {
			return true
		}
	}
	return false
}

//VLanEnsure creates a bridge (name) that is child of master (bridge) with vlan tag vlan
func VLanEnsure(name, bridge string, vlan uint16) (string, error) {
	//abstract method to ensure a bridge exists that has this vlan tag.
	if !BridgeExists(bridge) {
		return "", fmt.Errorf("master bridge does not exist")
	}

	portName := fmt.Sprintf("vlbr%dp", vlan)
	portPeerName := fmt.Sprintf("vlbr%din", vlan)

	if br, ok := PortToBridge(portName); ok {
		if br != bridge {
			return "", fmt.Errorf("reassigning vlang tag to another master bridge is not allowed")
		}
	}

	if br, ok := PortToBridge(portPeerName); ok {
		//peer already exists.
		if name == "" {
			return br, nil
		} else if br != name {
			return "", fmt.Errorf("reassigning vlan tag to another bridge not allowed")
		} else {
			//we already validated this setup.
			return name, nil
		}
	}

	if name == "" {
		name = fmt.Sprintf("vlbr%d", vlan)
	}

	if err := BridgeAdd(name); err != nil {
		return "", err
	}

	//add port in master
	if err := PortAdd(portName, bridge, vlan,
		TypeOption("patch"),
		PeerOption(portPeerName),
	); err != nil {
		return "", err
	}

	//connect port to vlan bridge
	if err := PortAdd(portPeerName, name, 0,
		TypeOption("patch"),
		PeerOption(portName),
	); err != nil {
		return "", err
	}

	return name, nil
}
