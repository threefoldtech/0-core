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

func (b *manager) getInputRuleHandler(set nft.Nft, rule string) int {
	filter, ok := set["filter"]
	if !ok {
		return -1
	}
	chain, ok := filter.Chains["input"]
	if !ok {
		return -1
	}

	for _, r := range chain.Rules {
		if r.Body == rule {
			return r.Handle
		}
	}

	return -1
}

func (b *manager) openPort(ctx pm.Context) (interface{}, error) {
	rule, err := b.parsePort(ctx)
	if err != nil {
		return nil, err
	}

	b.m.Lock()
	defer b.m.Unlock()

	ruleset, err := b.Get()
	if err != nil {
		return nil, err
	}

	handler := b.getInputRuleHandler(ruleset, rule)
	if handler >= 0 {
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

	ruleset, err := b.Get()
	if err != nil {
		return nil, err
	}
	handler := b.getInputRuleHandler(ruleset, rule)
	if handler == -1 {
		//nothing to do here
		return nil, nil
	}

	if err := b.Drop(nft.FamilyINET, "filter", "input", handler); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *manager) listPorts(ctx pm.Context) (interface{}, error) {
	b.m.RLock()
	defer b.m.RUnlock()

	ruleset, err := b.Get()
	if err != nil {
		return nil, err
	}

	filter, ok := ruleset["filter"]
	if !ok {
		return nil, nil
	}
	chain, ok := filter.Chains["input"]
	if !ok {
		return nil, nil
	}

	var rules []string
	for _, rule := range chain.Rules {
		rules = append(rules, rule.Body)
	}

	return rules, nil
}

func (b *manager) ruleExists(ctx pm.Context) (interface{}, error) {
	rule, err := b.parsePort(ctx)
	if err != nil {
		return nil, err
	}

	b.m.Lock()
	defer b.m.Unlock()

	ruleset, err := b.Get()
	if err != nil {
		return nil, err
	}

	handler := b.getInputRuleHandler(ruleset, rule)
	return handler >= 0, nil
}
