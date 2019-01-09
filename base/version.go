package base

import "fmt"

/*
The constants in this file are auto-replaced with the actual values
during the build of both core0 and coreX (only using the make file)
*/

var (
	Branch   = "{branch}"
	Revision = "{revision}"
	Dirty    = "{dirty}"
)

type Ver interface {
	Short() string
	String() string
}

type version struct {
	Branch   string `json:"branch"`
	Revision string `json:"revision"`
	Dirty    bool   `json:"dirty"`
}

func (v *version) String() string {
	s := fmt.Sprintf("Version: %s @Revision: %s", v.Branch, v.Revision)
	if Dirty != "" {
		s += " (dirty-repo)"
	}

	return s
}

func (v *version) Short() string {
	s := fmt.Sprintf("%s@%s", v.Branch, v.Revision[0:7])
	if Dirty != "" {
		s += "(D)"
	}
	return s
}

func Version() Ver {
	return &version{
		Branch:   Branch,
		Revision: Revision,
		Dirty:    len(Dirty) > 0,
	}
}
