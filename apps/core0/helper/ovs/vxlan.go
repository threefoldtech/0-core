package ovs

import (
	"fmt"
)

//VXLanEnsure ensures a vxlan bridge with name (name) that is sub to master bridge (bridge) with vxlan id (vxlan)
func VXLanEnsure(name, bridge string, vxlan uint) (string, error) {
	//abstract method to ensure a bridge exists that has this vlan tag.
	if !BridgeExists(bridge) {
		return "", fmt.Errorf("master bridge does not exist")
	}

	vtep, err := VtepEnsure(vxlan, bridge)

	if err != nil {
		return "", err
	}

	if br, ok := PortToBridge(vtep); ok {
		if name == "" {
			return br, nil
		} else if br != name {
			return "", fmt.Errorf("reassigning vxlan tag to another bridge not allowed")
		} else {
			return name, nil
		}
	}

	if name == "" {
		name = fmt.Sprintf("vxlbr%d", vxlan)
	}

	if err := BridgeAdd(name, nil); err != nil {
		return "", err
	}

	//add port in vxlan bridge
	if err := PortAdd(vtep, name, 0); err != nil {
		return "", err
	}

	return name, nil
}
