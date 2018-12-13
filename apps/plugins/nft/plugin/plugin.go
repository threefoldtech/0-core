package main

import (
	"github.com/threefoldtech/0-core/apps/plugins/nft"
	"github.com/threefoldtech/0-core/base/plugin"
)

func main() {} // silence the error

var (
	mgr manager
	_   nft.API = (*manager)(nil)

	Plugin = plugin.Plugin{
		Name:    "nft",
		Version: "1.0",
		Open: func(api plugin.API) error {
			mgr.api = api
			return nil
		},
		API: func() interface{} {
			return &mgr
		},
	}
)

type manager struct {
	api plugin.API
}
