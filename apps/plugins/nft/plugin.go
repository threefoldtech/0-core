package nft

import (
	"sync"

	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

func main() {} // silence the error

var (
	mgr manager
	_   API = (*manager)(nil)

	Plugin = plugin.Plugin{
		Name:      "nft",
		Version:   "1.0",
		CanUpdate: true,
		Open: func(api plugin.API) error {
			return newManager(&mgr, api)
		},
		API: func() interface{} {
			return &mgr
		},
		Actions: map[string]pm.Action{
			"open_port":   mgr.openPort,
			"drop_port":   mgr.dropPort,
			"list":        mgr.listPorts,
			"rule_exists": mgr.ruleExists,
		},
	}
)

var (
	nftInit = Nft{
		"nat": Table{
			Family: FamilyIP,
			Chains: Chains{
				"pre": Chain{
					Type:     TypeNAT,
					Hook:     "prerouting",
					Priority: 0,
					Policy:   "accept",
				},
				"post": Chain{
					Type:     TypeNAT,
					Hook:     "postrouting",
					Priority: 0,
					Policy:   "accept",
				},
			},
		},
		"filter": Table{
			Family: FamilyINET,
			Chains: Chains{
				"pre": Chain{
					Type:     TypeFilter,
					Hook:     "prerouting",
					Priority: 0,
					Policy:   "accept",
				},
				"input": Chain{
					Type:     TypeFilter,
					Hook:     "input",
					Priority: 0,
					Policy:   "drop",
					Rules: []Rule{
						{Body: "ct state {established, related} accept"},
						{Body: "iifname lo accept"},
						{Body: "iifname vxbackend accept"},
						{Body: "ip protocol icmp accept"},
					},
				},
				"forward": Chain{
					Type:     TypeFilter,
					Hook:     "forward",
					Priority: 0,
					Policy:   "accept",
				},
				"output": Chain{
					Type:     TypeFilter,
					Hook:     "output",
					Priority: 0,
					Policy:   "accept",
				},
			},
		},
	}
)

func newManager(mgr *manager, api plugin.API) error {
	mgr.api = api
	return mgr.init()
}

type manager struct {
	api plugin.API
	m   sync.RWMutex
}

func (m *manager) init() error {
	return m.Apply(nftInit)
}
