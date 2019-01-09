package plugin

import (
	"fmt"

	"github.com/threefoldtech/0-core/base/pm"
)

//Plugin description type
type Plugin struct {
	Name     string
	Version  string
	Requires []string
	Open     func(API) error
	Close    func() error
	API      func() interface{}
	Actions  map[string]pm.Action
}

func (p *Plugin) String() string {
	if len(p.Version) == 0 {
		return p.Name
	}

	return fmt.Sprintf("%s-%s", p.Name, p.Version)
}
