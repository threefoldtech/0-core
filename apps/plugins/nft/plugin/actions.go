package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/threefoldtech/0-core/apps/plugins/nft"
	"github.com/threefoldtech/0-core/base/pm"
)

type Port struct {
	Port      int    `json:"port"`
	Interface string `json:"interface,omitempty"`
	Subnet    string `json:"subnet,omitempty"`
}

func (b *manager) parsePort(ctx pm.Context) (string, error) {
	var args Port
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return "", err
	}

	body := ""
	if args.Interface != "" {
		body += fmt.Sprintf(`iifname "%s" `, args.Interface)
	}
	if args.Subnet != "" {
		subnet := args.Subnet
		_, net, err := net.ParseCIDR(args.Subnet)
		if err == nil {
			subnet = net.String()
		}

		body += fmt.Sprintf(`ip saddr %s `, subnet)
	}

	body += fmt.Sprintf(`tcp dport %d accept`, args.Port)

	return body, nil
}

func (b *manager) exists(rule string) bool {
	_, ok := b.api.Store().Get(rule)
	return ok
}

func (b *manager) register(rule string) error {
	if b.exists(rule) {
		return fmt.Errorf("exists")
	}

	b.api.Store().Set(rule, nil)
	return nil
}

func (b *manager) openPort(ctx pm.Context) (interface{}, error) {
	rule, err := b.parsePort(ctx)
	if err != nil {
		return nil, err
	}

	b.m.Lock()
	defer b.m.Unlock()

	if err := b.register(rule); err != nil {
		return nil, fmt.Errorf("rule exists")
	}

	n := nft.Nft{
		"filter": nft.Table{
			Family: nft.FamilyINET,
			Chains: nft.Chains{
				"input": nft.Chain{
					Rules: []nft.Rule{
						{Body: rule},
					},
				},
			},
		},
	}

	if err := b.Apply(n); err != nil {
		b.api.Store().Del(rule)
		return nil, err
	}

	return nil, nil
}

func (b *manager) dropPort(ctx pm.Context) (interface{}, error) {
	rule, err := b.parsePort(ctx)
	if err != nil {
		return nil, err
	}

	b.m.Lock()
	defer b.m.Unlock()

	if !b.exists(rule) {
		// nothing to do, just return
		return nil, nil
	}

	n := nft.Nft{
		"filter": nft.Table{
			Family: nft.FamilyINET,
			Chains: nft.Chains{
				"input": nft.Chain{
					Rules: []nft.Rule{
						{Body: rule},
					},
				},
			},
		},
	}

	if err := b.DropRules(n); err != nil {
		return nil, err
	}

	b.api.Store().Del(rule)
	return nil, nil
}

func (b *manager) listPorts(ctx pm.Context) (interface{}, error) {
	b.m.RLock()
	defer b.m.RUnlock()

	var ports []string
	for port := range b.api.Store().List() {
		ports = append(ports, port)
	}
	return ports, nil
}

func (b *manager) ruleExists(ctx pm.Context) (interface{}, error) {
	rule, err := b.parsePort(ctx)
	if err != nil {
		return nil, err
	}

	b.m.RLock()
	defer b.m.RUnlock()

	return b.exists(rule), nil
}
