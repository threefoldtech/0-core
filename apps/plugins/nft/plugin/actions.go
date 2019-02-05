package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/threefoldtech/0-core/apps/plugins/nft"
	"github.com/threefoldtech/0-core/base/pm"
)

type Port struct {
	Port      uint16 `json:"port"`
	Interface string `json:"interface,omitempty"`
	Subnet    string `json:"subnet,omitempty"`
}

func (b *manager) getArgs(ctx pm.Context) (args Port, err error) {
	cmd := ctx.Command()
	err = json.Unmarshal(*cmd.Arguments, &args)
	return
}

func (p *Port) getRule() string {
	body := ""
	if p.Interface != "" {
		body += fmt.Sprintf(`iifname "%s" `, p.Interface)
	}

	if p.Subnet != "" {
		subnet := p.Subnet
		_, net, err := net.ParseCIDR(p.Subnet)
		if err == nil {
			subnet = net.String()
		}

		body += fmt.Sprintf(`ip saddr %s `, subnet)
	}

	body += fmt.Sprintf(`tcp dport %d accept`, p.Port)

	return body
}

func (b *manager) openPort(ctx pm.Context) (interface{}, error) {
	args, err := b.getArgs(ctx)
	if err != nil {
		return nil, err
	}

	b.m.Lock()
	defer b.m.Unlock()

	matches, err := b.Find(nft.And{
		&nft.TableFilter{Table: "filter"},
		&nft.ChainFilter{Chain: "input"},
		&nft.IntMatchFilter{Name: "tcp", Field: "dport", Value: uint64(args.Port)},
	})

	if err != nil {
		return nil, err
	}

	if len(matches) != 0 {
		return nil, fmt.Errorf("rule already exists for port: %d", args.Port)
	}

	n := nft.Nft{
		"filter": nft.Table{
			Family: nft.FamilyINET,
			Chains: nft.Chains{
				"input": nft.Chain{
					Rules: []nft.Rule{
						{Body: args.getRule()},
					},
				},
			},
		},
	}

	if err := b.Apply(n); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *manager) dropPort(ctx pm.Context) (interface{}, error) {
	args, err := b.getArgs(ctx)
	if err != nil {
		return nil, err
	}

	b.m.Lock()
	defer b.m.Unlock()

	matches, err := b.Find(nft.And{
		&nft.TableFilter{Table: "filter"},
		&nft.ChainFilter{Chain: "input"},
		&nft.IntMatchFilter{Name: "tcp", Field: "dport", Value: uint64(args.Port)},
	})

	if err != nil {
		return nil, err
	}

	for _, rule := range matches {
		if err := b.Drop(nft.FamilyINET, "filter", "input", rule.Handle); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (b *manager) listPorts(ctx pm.Context) (interface{}, error) {
	args, err := b.getArgs(ctx)
	if err != nil {
		return nil, err
	}

	b.m.Lock()
	defer b.m.Unlock()

	matches, err := b.Find(nft.And{
		&nft.TableFilter{Table: "filter"},
		&nft.ChainFilter{Chain: "input"},
		&nft.IntMatchFilter{Name: "tcp", Field: "dport", Value: uint64(args.Port)},
	})

	if err != nil {
		return nil, err
	}

	var rules []string
	for _, rule := range matches {
		rules = append(rules, rule.Body)
	}

	return rules, nil
}

func (b *manager) ruleExists(ctx pm.Context) (interface{}, error) {
	args, err := b.getArgs(ctx)
	if err != nil {
		return nil, err
	}

	b.m.Lock()
	defer b.m.Unlock()

	matches, err := b.Find(nft.And{
		&nft.TableFilter{Table: "filter"},
		&nft.ChainFilter{Chain: "input"},
		&nft.IntMatchFilter{Name: "tcp", Field: "dport", Value: uint64(args.Port)},
	})

	if err != nil {
		return nil, err
	}

	return len(matches) > 0, nil
}
