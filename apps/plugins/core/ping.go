package core

import (
	"fmt"

	"github.com/threefoldtech/0-core/base/pm"
)

func (mgr *coreManager) ping(ctx pm.Context) (interface{}, error) {
	return fmt.Sprintf("PONG %s", mgr.api.Version()), nil
}
