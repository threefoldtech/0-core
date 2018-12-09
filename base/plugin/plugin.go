package plugin

import (
	"github.com/threefoldtech/0-core/base/pm"
)

//Plugin description type
type Plugin struct {
	Name     string
	Version  string //float?
	Requires []string
	Open     func(API) error
	Close    func() error
	Actions  map[string]pm.Action
}
