package containers

import "github.com/threefoldtech/0-core/base/pm"

//Container api
type Container interface {
	ID() uint16
	Arguments() (ContainerCreateArguments, error)
	Root() string
}

//API defines container plugin api
type API interface {
	Dispatch(id uint16, cmd *pm.Command) (*pm.JobResult, error)
	GetWithTags(tags ...string) []Container
	GetOneWithTags(tags ...string) Container
	Of(id uint16) Container
}
