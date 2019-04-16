package containers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/threefoldtech/0-core/base/pm"
)

func (m *Manager) ztInfo(ctx pm.Context) (interface{}, error) {
	var args ContainerArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	cont := loadContainer(m, args.Container)

	job, err := m.api.System(
		"ip", "netns", "exec", fmt.Sprintf("%d", args.Container),
		"zerotier-cli", "-j", fmt.Sprintf("-D%s", cont.zerotierHome()), "info",
	)

	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(job.Streams.Stdout()), &data); err != nil {
		return nil, err
	}

	//inject private identity
	secret, err := ioutil.ReadFile(path.Join(cont.zerotierHome(), "identity.secret"))
	data["secretIdentity"] = strings.TrimSpace(string(secret))

	return data, nil
}

func (m *Manager) ztList(ctx pm.Context) (interface{}, error) {
	var args ContainerArguments
	cmd := ctx.Command()
	if err := json.Unmarshal(*cmd.Arguments, &args); err != nil {
		return nil, err
	}

	cont := loadContainer(m, args.Container)

	job, err := m.api.System(
		"ip", "netns", "exec", fmt.Sprintf("%d", args.Container),
		"zerotier-cli", "-j", fmt.Sprintf("-D%s", cont.zerotierHome()), "listnetworks",
	)

	if err != nil {
		return nil, err
	}

	var data interface{}
	if err := json.Unmarshal([]byte(job.Streams.Stdout()), &data); err != nil {
		return nil, err
	}

	return data, nil
}
