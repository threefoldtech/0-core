package socat

import (
	"fmt"
)

const (
	//Container system
	Container uint8 = iota + 1
	//KVM system
	KVM
)

type API interface {
	SetPortForward(ns NS, ip string, host string, dest int) error
	RemovePortForward(ns NS, host string, dest int) error
	RemoveAll(ns NS) error
	List(ns NS) (PortMap, error)
	ListAll(system uint8) (map[NS]PortMap, error)
	Resolve(address string) string
	ResolveURL(raw string) (string, error)
	ValidHost(host string) bool
}

type NS uint32

func (n NS) String() string {
	return fmt.Sprintf("%08x", uint32(n))
}

func Namespace(system uint8, id uint16) NS {
	return NS(uint32(system)<<(8*3) | uint32(id))
}

type PortMap map[string]int
