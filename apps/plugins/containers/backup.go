package containers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"regexp"
	"syscall"

	"github.com/threefoldtech/0-core/base/pm"
)

const (
	backupMetaName = ".corex.meta"
)

var (
	resticSnaphostIdP = regexp.MustCompile(`snapshot ([^\s]+) saved`)
)

func (m *Manager) backup(ctx pm.Context) (interface{}, error) {
	var args struct {
		Container uint16   `json:"container"`
		URL       string   `json:"url"`
		Tags      []string `json:"tags"`
	}
	cmd := ctx.Command()

	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	if args.Container <= 0 {
		return nil, fmt.Errorf("invalid container id")
	}

	cont := loadContainer(m, args.Container)

	//pause container
	//TODO: avoid race if cont has just started and pid is not set yet!
	if cont.PID == 0 {
		return nil, fmt.Errorf("container is not fully started yet")
	}

	u, err := url.Parse(args.URL)
	if err != nil {
		return nil, err
	}

	password := u.Query().Get("password")
	u.Fragment = "" //just to make sure
	repo := args.URL
	if u.Scheme == "file" || len(u.Scheme) == 0 {
		repo = u.Path
	} else {
		u.RawQuery = ""
		repo = u.String()
	}

	restic := []string{
		"-r", repo,
		"backup",
		"--exclude", "proc/**",
		"--exclude", "dev/**",
		"--exclude", "sys/**",
	}
	arguments, err := cont.Arguments()
	if err != nil {
		return nil, err
	}
	for _, tag := range arguments.Tags {
		restic = append(restic, "--tag", tag)
	}

	for _, tag := range args.Tags {
		restic = append(restic, "--tag", tag)
	}

	root := cont.Root()

	//write meta
	var nics []*Nic
	for _, n := range arguments.Nics {
		if n.State == NicStateConfigured {
			nics = append(nics, n)
		}
	}
	arguments.Nics = nics
	mf := path.Join(root, backupMetaName)
	meta, err := json.Marshal(arguments)
	if err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(mf, meta, 0400); err != nil {
		return nil, err
	}

	defer os.Remove(mf)

	//we specify files to backup one by one instead of a full dire to
	//have more control
	items, err := ioutil.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, item := range items {
		if item.Name() == "coreX" {
			continue
		}

		files = append(files, path.Join(root, item.Name()))
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("nothing to backup")
	}

	restic = append(restic, files...)

	//pause container
	syscall.Kill(-cont.PID, syscall.SIGSTOP)
	defer syscall.Kill(-cont.PID, syscall.SIGCONT)

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
		return nil, err
	}

	result := job.Wait()
	if result.State != pm.StateSuccess {
		return nil, fmt.Errorf("failed to backup container: %s", result.Streams.Stderr())
	}

	//read snapshot id
	match := resticSnaphostIdP.FindStringSubmatch(result.Streams.Stdout())
	if len(match) != 2 {
		return nil, fmt.Errorf("failed to retrieve snapshot ID")
	}

	return match[1], nil
}

func (m *Manager) restore(ctx pm.Context) (interface{}, error) {
	var args struct {
		URL string `json:"url"`
	}
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	tmp, err := ioutil.TempDir("", "restic")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(tmp)

	if err := m.filesystem().RestoreRepo(args.URL, tmp, backupMetaName); err != nil {
		return nil, err
	}

	meta, err := os.Open(path.Join(tmp, backupMetaName))
	if err != nil {
		return nil, err
	}

	defer meta.Close()

	dec := json.NewDecoder(meta)

	var cargs ContainerCreateArguments
	if err := dec.Decode(&cargs); err != nil {
		return nil, err
	}

	//set restore url
	//rewrite the URL to use restic prefix. now we can call create.
	cargs.Root = fmt.Sprintf("restic:%s", args.URL)
	cargs.Tags = cmd.Tags //override original tags

	cont, err := m.createContainer(cargs)
	if err != nil {
		return nil, err
	}

	return cont.id, nil
}
