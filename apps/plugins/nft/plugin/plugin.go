package main

import (
	"sync"

	"github.com/threefoldtech/0-core/apps/plugins/nft"
	"github.com/threefoldtech/0-core/base/plugin"
	"github.com/threefoldtech/0-core/base/pm"
)

func main() {} // silence the error

var (
	mgr *manager
	_   nft.API = (*manager)(nil)

	Plugin = plugin.Plugin{
		Name:    "nft",
		Version: "1.0",
		Open: func(api plugin.API) (err error) {
			mgr, err = newManager(api)
			if err != nil {
				return
			}

			return mgr.init()
		},
		API: func() interface{} {
			return mgr
		},
		Actions: map[string]pm.Action{
			"open_port": nil,
		},
	}
)

var (
	nftInit = nft.Nft{
		"nat": nft.Table{
			Family: nft.FamilyIP,
			Chains: nft.Chains{
				"pre": nft.Chain{
					Type:     nft.TypeNAT,
					Hook:     "prerouting",
					Priority: 0,
					Policy:   "accept",
				},
				"post": nft.Chain{
					Type:     nft.TypeNAT,
					Hook:     "postrouting",
					Priority: 0,
					Policy:   "accept",
				},
			},
		},
		"filter": nft.Table{
			Family: nft.FamilyINET,
			Chains: nft.Chains{
				"pre": nft.Chain{
					Type:     nft.TypeFilter,
					Hook:     "prerouting",
					Priority: 0,
					Policy:   "accept",
				},
				"input": nft.Chain{
					Type:     nft.TypeFilter,
					Hook:     "input",
					Priority: 0,
					Policy:   "drop",
					Rules: []nft.Rule{
						{Body: "ct state {established, related} accept"},
						{Body: "iifname lo accept"},
						{Body: "iifname vxbackend accept"},
						{Body: "ip protocol icmp accept"},
					},
				},
				"forward": nft.Chain{
					Type:     nft.TypeFilter,
					Hook:     "forward",
					Priority: 0,
					Policy:   "accept",
				},
				"output": nft.Chain{
					Type:     nft.TypeFilter,
					Hook:     "output",
					Priority: 0,
					Policy:   "accept",
				},
			},
		},
	}
)

func newManager(api plugin.API) (*manager, error) {
	return &manager{
		api:   api,
		rules: make(map[string]struct{}),
	}, nil
}

type manager struct {
	api plugin.API

	rules map[string]struct{}
	m     sync.RWMutex
}

func (m *manager) init() error {
	return m.Apply(nftInit)
}
