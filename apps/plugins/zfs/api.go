package zfs

import "github.com/threefoldtech/0-core/base/pm"

type API interface {
	MountFList(namespace, storage, src string, target string, hooks ...pm.RunnerHook) error
	MergeFList(namespace, target, base, flist string) error
	RestoreRepo(repo, target string, include ...string) error
	GetCacheZeroFSDir() string
}
