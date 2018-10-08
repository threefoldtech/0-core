package ovs

import (
	"fmt"
)

//BondAdd add bond port with name (name) to bridge (bridge)
func BondAdd(name, bridge string, mode BondMode, lacp bool, links ...string) error {
	cmd := []string{"add-bond", bridge, name}
	cmd = append(cmd, links...)
	if lacp {
		cmd = append(cmd, "lacp=active")
	}

	cmd = append(cmd, fmt.Sprintf("bond_mode=%v", mode))

	_, err := vsctl(cmd...)
	return err
}
