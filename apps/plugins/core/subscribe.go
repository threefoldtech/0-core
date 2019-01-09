package core

import (
	"encoding/json"
	"fmt"

	"github.com/threefoldtech/0-core/base/pm"
	"github.com/threefoldtech/0-core/base/stream"
)

func (mgr *coreManager) subscribe(ctx pm.Context) (interface{}, error) {
	var args struct {
		ID string `json:"id"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	job, ok := mgr.api.JobOf(args.ID)

	if !ok {
		return nil, fmt.Errorf("job '%s' does not exist", args.ID)
	}

	job.Subscribe(func(msg *stream.Message) {
		ctx.Message(msg)
	})

	job.Wait()
	return nil, nil
}
