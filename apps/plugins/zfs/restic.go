package zfs

import (
	"fmt"
	"net/url"

	"github.com/threefoldtech/0-core/base/pm"
)

func (m *Manager) RestoreRepo(repo, target string, include ...string) error {
	//file://password/path/to/repo
	u, err := url.Parse(repo)
	if err != nil {
		return err
	}

	password := u.Query().Get("password")
	snapshot := u.Fragment
	if u.Scheme == "file" || len(u.Scheme) == 0 {
		repo = u.Path
	} else {
		u.Fragment = ""
		u.RawQuery = ""
		repo = u.String()
	}

	restic := []string{
		"-r", repo,
		"restore",
		"-t", target,
	}

	for _, i := range include {
		restic = append(restic, "-i", i)
	}

	restic = append(restic, snapshot)

	job, err := m.api.Run(
		&pm.Command{
			Command: pm.CommandSystem,
			Arguments: pm.MustArguments(
				pm.SystemCommandArguments{
					Name:  "restic",
					Args:  restic,
					StdIn: password,
				},
			),
		},
	)

	if err != nil {
		return err
	}

	if result := job.Wait(); result.State != pm.StateSuccess {
		return fmt.Errorf("failed to restore snapshot: %s", result.Streams.Stderr())
	}

	return nil
}
