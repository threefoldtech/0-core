package protocol

import "github.com/threefoldtech/0-core/base/pm"

//Database interface to low level 0-core store (usually redis)
type Database interface {
	GetNext(queue string, cmd *pm.Command) error
	Set(result *pm.JobResult) error
	Get(id string, timeout int) (*pm.JobResult, error)
	Flag(id string) error
	UnFlag(id string) error
	Flagged(id string) bool
}
