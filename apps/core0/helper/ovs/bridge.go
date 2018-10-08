package ovs

import (
	"strings"
)

//BridgeExists checks if bridge exists
func BridgeExists(name string) bool {
	if _, err := vsctl("br-exists", name); err != nil {
		return false
	}

	return true
}

//BridgeAdd add a new bridge
func BridgeAdd(name string, options ...Option) error {
	commands := []string{"add-br", name}
	for _, opt := range options {
		commands = append(commands, "--", "set", "Bridge", name, opt.String())
	}
	_, err := vsctl(commands...)
	return err
}

//BridgeDel del a bridge
func BridgeDel(name string) error {
	_, err := vsctl("del-br", name)
	return err
}

//BridgeList list all bridges
func BridgeList() ([]string, error) {
	out, err := vsctl("list-br")
	if err != nil {
		return nil, err
	}

	return strings.Fields(out), nil
}
