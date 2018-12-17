package plugin

import "github.com/threefoldtech/0-core/base/pm"

//API defines api entry point for plugins
type API interface {
	Run(cmd *pm.Command, hooks ...pm.RunnerHook) (pm.Job, error)
	System(bin string, args ...string) (*pm.JobResult, error)
	Internal(cmd string, args pm.M, out interface{}) error
	JobOf(id string) (pm.Job, bool)
	Jobs() map[string]pm.Job
	Plugin(name string) (interface{}, error)
}
