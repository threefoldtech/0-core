package plugin

import (
	"github.com/threefoldtech/0-core/base"
	"github.com/threefoldtech/0-core/base/pm"
)

//API defines api entry point for plugins
type API interface {
	Version() base.Ver
	Run(cmd *pm.Command, hooks ...pm.RunnerHook) (pm.Job, error)
	System(bin string, args ...string) (*pm.JobResult, error)
	Internal(cmd string, args pm.M, out interface{}) error
	JobOf(id string) (pm.Job, bool)
	Jobs() map[string]pm.Job
	Plugin(name string) (interface{}, error)
	Shutdown(except ...string)
	Aggregate(op, key string, value float64, id string, tags ...pm.Tag)
	Store() Store
}

//Store stores data on the core0 context
type Store interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	List() (map[string][]byte, error)
}
