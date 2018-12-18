package core

import (
	"fmt"

	"github.com/threefoldtech/0-core/base/pm"
)

func ping(ctx pm.Context) (interface{}, error) {
	return fmt.Sprintf("PONG %s", api.Version()), nil
}
