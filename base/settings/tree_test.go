package settings

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

var (
	settings IncludedSettings = IncludedSettings{
		Startup: map[string]Startup{
			"ovc": Startup{
				After: []string{"mongo", "influx"},
			},
			"fstab": Startup{
				After: []string{string(AfterInit), "udev"},
			},
			"mongo": Startup{
				After: []string{string(AfterBoot)},
			},
			"sshd": Startup{
				After: []string{string(AfterNet)},
			},
			"influx": Startup{
				After: []string{string(AfterBoot)},
			},
			"udev": Startup{
				After: []string{string(AfterInit)},
			},
		},
	}
)

func TestGetTree(t *testing.T) {
	tree, errors := settings.GetStartupTree()
	if ok := assert.Empty(t, errors); !ok {
		t.Fail()
	}

	if ok := assert.NotNil(t, tree); !ok {
		t.Fatal()
	}
}

func AssertTree(t *testing.T, expected [][]string, tree StartupTree) bool {
	services := tree.Services()
	offset := 0
	ok := true
	for i, e := range expected {
		slice := services[offset : offset+len(e)]
		actual := make([]string, 0)
		for _, s := range slice {
			actual = append(actual, s.Key())
		}

		sort.Strings(e)
		sort.Strings(actual)

		if !assert.Equal(t, e, actual) {
			t.Fail()
			ok = false
		}
		offset = i + len(e)
	}

	return ok
}

func TestGetTreeSorted(t *testing.T) {
	tree, errors := settings.GetStartupTree()
	if ok := assert.Empty(t, errors); !ok {
		t.Fail()
	}

	if ok := assert.NotNil(t, tree); !ok {
		t.Fatal()
	}

	//expected order
	expected := [][]string{
		[]string{"udev"},
		[]string{"fstab"},
		[]string{"sshd"},
		[]string{"mongo", "influx"},
		[]string{"ovc"}}

	if !AssertTree(t, expected, tree) {
		t.Fatal()
	}
}

func TestTreeSlice(t *testing.T) {
	tree, errors := settings.GetStartupTree()
	if ok := assert.Empty(t, errors); !ok {
		t.Fail()
	}

	if ok := assert.NotNil(t, tree); !ok {
		t.Fatal()
	}

	slice := tree.Slice(Priority[AfterInit], Priority[AfterNet])
	if ok := assert.Len(t, slice, 2); !ok {
		t.Fatal()
	}

	expected := []string{"fstab", "udev"}
	actual := make([]string, 0)
	for _, s := range slice {
		actual = append(actual, s.Key())
	}
	sort.Strings(expected)
	sort.Strings(actual)

	if ok := assert.Equal(t, expected, actual); !ok {
		t.Fatal()
	}
}

func TestTreeSliceBoot(t *testing.T) {
	tree, errors := settings.GetStartupTree()
	if ok := assert.Empty(t, errors); !ok {
		t.Fail()
	}

	if ok := assert.NotNil(t, tree); !ok {
		t.Fatal()
	}

	slice := tree.Slice(Priority[AfterBoot], -1)
	if ok := assert.Len(t, slice, 3); !ok {
		t.Fatal()
	}

	expected := []string{"mongo", "influx", "ovc"}
	actual := make([]string, 0)
	for _, s := range slice {
		actual = append(actual, s.Key())
	}
	sort.Strings(expected)
	sort.Strings(actual)

	if ok := assert.Equal(t, expected, actual); !ok {
		t.Fatal()
	}
}

func TestGetTreeDefaultWeight(t *testing.T) {
	settings := IncludedSettings{
		Startup: map[string]Startup{
			"mount": Startup{
				After: []string{string(AfterInit)},
			},
			"mongo": Startup{
				After: []string{},
			},
			"influx": Startup{
				After: []string{},
			},
		},
	}

	tree, errors := settings.GetStartupTree()
	if ok := assert.Empty(t, errors); !ok {
		t.Fail()
	}

	if ok := assert.NotNil(t, tree); !ok {
		t.Fatal()
	}

	//expected order
	expected := [][]string{
		[]string{"mount"},
		[]string{"mongo", "influx"},
	}

	if !AssertTree(t, expected, tree) {
		t.Fatal()
	}
}

func TestGetTreeMissingDependency(t *testing.T) {
	settings := IncludedSettings{
		Startup: map[string]Startup{
			"mongo": Startup{
				After: []string{string(AfterBoot)},
			},
			"ovc": Startup{
				After: []string{"mongo", "influx"},
			},
		},
	}

	tree, errors := settings.GetStartupTree()
	if ok := assert.NotEmpty(t, errors); !ok {
		t.Fail()
	}

	if ok := assert.Len(t, tree.Services(), 1); !ok {
		t.Fatal()
	}
}

func TestGetTreeCyclicDependency(t *testing.T) {
	settings := IncludedSettings{
		Startup: map[string]Startup{
			"independent": Startup{
				After: []string{},
			},
			"influx": Startup{
				After: []string{},
			},
			"mongo": Startup{
				After: []string{"ovc"},
			},
			"ovc": Startup{
				After: []string{"mongo", "influx"},
			},
		},
	}

	tree, errors := settings.GetStartupTree()
	if ok := assert.NotEmpty(t, errors); !ok {
		t.Fail()
	}

	if ok := assert.Len(t, tree.Services(), 2); !ok {
		t.Fatal()
	}
}

func TestOVSDependencies(t *testing.T) {
	settings := IncludedSettings{
		Startup: map[string]Startup{
			"ovs.init": Startup{},
			"ovsdb": Startup{
				After: []string{"ovs.init"},
			},
			"ovs": Startup{
				After: []string{"ovsdb"},
			},
		},
	}

	_, errs := settings.GetStartupTree()
	if ok := assert.Empty(t, errs); !ok {
		t.Fatal()
	}
}
