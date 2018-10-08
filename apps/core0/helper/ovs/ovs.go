package ovs

import (
	"fmt"

	"github.com/threefoldtech/0-core/base/pm"
)

const (
	Binary = "ovs-vsctl"
)

type Option interface {
	fmt.Stringer
}

type KeyValueOption struct {
	Key   string
	Value string
}

func (kv KeyValueOption) String() string {
	return fmt.Sprintf("\"%s=%s\"", kv.Key, kv.Value)
}

func TypeOption(t string) Option {
	return KeyValueOption{Key: "type", Value: t}
}

func PeerOption(p string) Option {
	return KeyValueOption{Key: "option:peer", Value: p}
}

func HWAddrOption(mac string) Option {
	return KeyValueOption{Key: "other-config:hwaddr", Value: mac}
}

func MakeOptions(m map[string]string) []Option {
	var options []Option
	for k, v := range m {
		options = append(options, KeyValueOption{k, v})
	}

	return options
}

func vsctl(args ...string) (string, error) {
	job, err := pm.System(Binary, args...)
	if err != nil {
		return "", err
	}

	return job.Streams.Stdout(), nil
}

func set(table, record string, option ...Option) error {
	args := []string{"set", table, record}
	for _, opt := range option {
		args = append(args, opt.String())
	}

	_, err := vsctl(args...)
	return err
}

type BondMode string

const (
	BondModeActiveBackup = BondMode("active-backup")
	BondModeBalanceSLB   = BondMode("balance-slb")
	BondModeBalanceTCP   = BondMode("balance-tcp")
)

type BondAddArguments struct {
	Bridge
	Port  string   `json:"port"`
	Links []string `json:"links"`
	Mode  BondMode `json:"mode"`
	LACP  bool     `json:"lacp"`
}

func (b *BondAddArguments) Validate() error {
	if err := b.Bridge.Validate(); err != nil {
		return err
	}

	if b.Port == "" {
		return fmt.Errorf("missing port name")
	}

	if len(b.Links) <= 1 {
		return fmt.Errorf("need more than one link to bond")
	}

	return nil
}
