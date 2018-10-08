package ovs

import (
	"fmt"
	"strings"
)

//PortAdd adds a port
func PortAdd(name, bridge string, vlan uint16, option ...Option) error {
	var err error
	if vlan == 0 {
		_, err = vsctl("add-port", bridge, name)
	} else {
		_, err = vsctl("add-port", bridge, name, fmt.Sprintf("tag=%d", vlan))
	}

	if err != nil {
		return err
	}

	//setting options
	if len(option) != 0 {
		return set("Interface", name, option...)
	}

	return nil
}

//PortDel deletes a port
func PortDel(name, bridge string) error {
	var err error
	if bridge == "" {
		_, err = vsctl("del-port", name)
	} else {
		_, err = vsctl("del-port", bridge, name)
	}

	return err
}

//PortList list ports on a bridge
func PortList(br string) ([]string, error) {
	output, err := vsctl("list-ports", br)
	if err != nil {
		return nil, err
	}

	return strings.Fields(output), nil
}

//PortToBridge get the bridge which has this port attached
func PortToBridge(port string) (string, bool) {
	out, err := vsctl("port-to-br", port)
	if err != nil {
		return "", false
	}

	return strings.TrimSpace(out), true
}
