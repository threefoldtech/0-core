package builtin

import "github.com/Zero-OS/0-Core/base/pm"

func init() {
	pm.RegisterCmd("bash", "sh", "", []string{"-c", "{script}"}, nil)
}
