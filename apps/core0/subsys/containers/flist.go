package containers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"syscall"

	"github.com/threefoldtech/0-core/base/pm"
)

const (
	cmdFlistCreate = "corex.flist.create"
)

func zflist(args ...string) (*pm.JobResult, error) {
	log.Debugf("zflist %v", args)
	return pm.System("zflist", args...)
}

func containerPath(container *container, path string) string {
	return filepath.Join(container.Root, path)
}

type createArgs struct {
	Container uint16 `json:"container"`
	Flist     string `json:"flist"`   //path where to create the flist
	Storage   string `json:"storage"` // zdb://host:port to the data storage
	Src       string `json:"src"`     //path to the directory to create flist from
}

func (c createArgs) Validate() error {
	if c.Container <= 0 {
		return fmt.Errorf("invalid container id")
	}
	if c.Flist == "" {
		return fmt.Errorf("flist destination need to be specified")
	}
	if c.Storage == "" {
		return fmt.Errorf("flist data storage need to be specified")
	}
	if c.Src == "" {
		return fmt.Errorf("source directory need to be specified")
	}
	return nil
}

func (m *containerManager) flistCreate(cmd *pm.Command) (interface{}, error) {
	var args createArgs

	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	if err := args.Validate(); err != nil {
		return nil, err
	}

	m.conM.RLock()
	cont, ok := m.containers[args.Container]
	m.conM.RUnlock()

	if !ok {
		return nil, fmt.Errorf("container does not exist")
	}

	//pause container
	//TODO: avoid race if cont has just started and pid is not set yet!
	if cont.PID == 0 {
		return nil, fmt.Errorf("container is not fully started yet")
	}

	//pause container
	syscall.Kill(-cont.PID, syscall.SIGSTOP)
	defer syscall.Kill(-cont.PID, syscall.SIGCONT)

	archivePath := containerPath(cont, args.Flist)
	_, err := zflist("--archive", archivePath, "--create", containerPath(cont, args.Src), "--backend", args.Storage)
	return archivePath, err
}
