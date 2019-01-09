package protocol

import "github.com/threefoldtech/0-core/base/pm"

//API interface to have access to job results and state
type API interface {
	Set(result *pm.JobResult) error
	Get(id string, timeout int) (*pm.JobResult, error)
	Flag(id string) error
	UnFlag(id string) error
	Flagged(id string) bool
	//Database provides a low level access to the data store
	Database() Database
}

//Database interface to have very low level access to 0-core kv store (redis)
type Database interface {
	RPush(key string, args ...[]byte) (int64, error)
	LTrim(key string, start, stop int64) error
	GetKey(key string) ([]byte, error)
	SetKey(key string, value []byte) error
	DelKeys(keys ...string) (int64, error)
	LExpire(key string, duration int64) (int64, error)
}
