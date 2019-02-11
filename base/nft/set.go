package nft

import (
	"fmt"
	"strings"

	"github.com/threefoldtech/0-core/base/pm"
)

//IPv4Set creates/updates element set of type ipv4_addr
func IPv4Set(family Family, table string, name string, ips ...string) error {
	//nft add set ip nat host { type ipv4_addr\; }
	//nft add element ip nat host { 172^C9.0.1, 172.18.0.1 }

	_, err := pm.System("nft", "add", "set", string(family), table, name, "{", "type", "ipv4_addr;", "}")
	if err != nil {
		return err
	}

	if len(ips) == 0 {
		return nil
	}

	s := strings.Join(ips, ", ")
	_, err = pm.System("nft", "add", "element", string(family), table, name, "{", s, "}")

	return err
}

//IPv4SetDel delete ips from a ipv4_addr set
func IPv4SetDel(family Family, table, name string, ips ...string) error {
	if len(ips) == 0 {
		return nil
	}

	s := strings.Join(ips, ", ")
	_, err := pm.System("nft", "delete", "element", string(family), table, name, "{", s, "}")

	return err
}

//IPv4SetGet gets the current ipv4 set
func IPv4SetGet(family Family, table, name string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
