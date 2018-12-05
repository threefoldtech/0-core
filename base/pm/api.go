package pm

//API defines api entry point for plugins
type API interface {
	Run(cmd *Command, hooks ...RunnerHook) (Job, error)
	System(bin string, args ...string) (*JobResult, error)
	Internal(cmd string, args M, out interface{}) error
	JobOf(id string) (Job, bool)
}
