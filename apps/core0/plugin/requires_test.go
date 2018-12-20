package plugin

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/threefoldtech/0-core/base/plugin"
)

func TestRequires(t *testing.T) {
	p := plugins{
		&plugin.Plugin{
			Name:     "require_1",
			Requires: []string{"nodep"},
		},
		&plugin.Plugin{
			Name:     "require_2",
			Requires: []string{"nodep", "require_1"},
		},
		&plugin.Plugin{
			Name:     "cyclic_1",
			Requires: []string{"cyclic_2"},
		},
		&plugin.Plugin{
			Name:     "cyclic_2",
			Requires: []string{"cyclic_1"},
		},
		&plugin.Plugin{
			Name: "nodep",
		},
	}

	/*
		Currently we don't handle cyclic dependency we just makeing sure
		it will not cause fatal crashes to the system.
	*/
	sort.Sort(p)

	if ok := assert.Len(t, p, 5); !ok {
		t.Fatal()
	}

	var names []string
	for _, pl := range p {
		names = append(names, pl.Name)
	}

	if ok := assert.Equal(t, []string{"nodep", "require_1", "require_2", "cyclic_2", "cyclic_1"}, names); !ok {
		t.Error()
	}
}
