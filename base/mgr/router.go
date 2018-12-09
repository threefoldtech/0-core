package mgr

import (
	"github.com/threefoldtech/0-core/base/pm"
)

//Router defines a command router
type Router interface {
	Get(name string) (pm.Action, error)
}
