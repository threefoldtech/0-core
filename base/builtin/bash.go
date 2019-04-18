package builtin

import "github.com/threefoldtech/0-core/base/pm"

func init() {
	pm.RegisterExtension("bash", "sh", "", []string{"-c", "{script}"}, nil)
}
