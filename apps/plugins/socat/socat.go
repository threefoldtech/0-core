package socat

import "fmt"

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
	Resolve(address string) string
	ResolveURL(raw string) (string, error)
	ValidHost(host string) bool
}

type NS uint32

func (n NS) String() string {
	return fmt.Sprintf("%08x", uint32(n))
}

func Namespace(subsystem uint8, id uint16) NS {
	return NS(uint32(subsystem)<<(8*3) | uint32(id))
}
