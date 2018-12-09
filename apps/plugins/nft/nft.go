package main

import (
	"github.com/threefoldtech/0-core/base/plugin"
)

var (
	nft manager
	_   API = (*manager)(nil)

	Plugin = plugin.Plugin{
		Name:    "nft",
		Version: "1.0",
		Open: func(api plugin.API) error {
			nft.api = api
			return nil
		},
		API: func() interface{} {
			return &nft
		},
	}
)

type manager struct {
	api plugin.API
}

//API defines nft api
type API interface {
	ApplyFromFile(cfg string) error
	Apply(nft Nft) error
	DropRules(sub Nft) error
	Drop(family Family, table, chain string, handle int) error
	Get() (Nft, error)

	IPv4Set(family Family, table string, name string, ips ...string) error
	IPv4SetDel(family Family, table, name string, ips ...string) error
}
