package socat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/0-core/apps/plugins/nft"
)

func TestResolveInvalidSyntax(t *testing.T) {
	api := testNFT{
		rules: []nft.FilterRule{
			{
				Rule: nft.Rule{
					Body: "dnat to 1.2.3.4:80",
				},
			},
		},
	}
	address := "1234"
	mgr := &socatManager{}

	if ok := assert.Equal(t, address, mgr.resolve(address, &api)); !ok {
		t.Error()
	}

	address = ":1234"
	if ok := assert.Equal(t, address, mgr.resolve(address, &api)); !ok {
		t.Error()
	}
}

type testNFT struct {
	rules []nft.FilterRule
}

func (t *testNFT) Apply(nft nft.Nft) error {
	panic("not implemented")
}
func (t *testNFT) Drop(family nft.Family, table, chain string, handle int) error {
	panic("not implemented")
}
func (t *testNFT) Find(filter ...nft.Filter) ([]nft.FilterRule, error) {
	return t.rules, nil
}

func (t *testNFT) IPv4Set(family nft.Family, table string, name string, ips ...string) error {
	panic("not implemented")
}
func (t *testNFT) IPv4SetDel(family nft.Family, table, name string, ips ...string) error {
	panic("not implemented")
}

func TestResolve(t *testing.T) {
	api := testNFT{
		rules: []nft.FilterRule{
			{
				Rule: nft.Rule{
					Body: "dnat to 1.2.3.4:80",
				},
			},
		},
	}

	mgr := &socatManager{}

	if ok := assert.Equal(t, "1.2.3.4:80", mgr.resolve("127.0.0.1:8080", &api)); !ok {
		t.Error()
	}
}

func TestResolveURL(t *testing.T) {
	api := testNFT{
		rules: []nft.FilterRule{
			{
				Rule: nft.Rule{
					Body: "dnat to 1.2.3.4:8080",
				},
			},
		},
	}

	mgr := &socatManager{}

	url, err := mgr.resolveURL("http://127.0.0.1/", &api)

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "http://1.2.3.4:8080/", url); !ok {
		t.Error()
	}

	url, err = mgr.resolveURL("http://localhost/", &api)

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "http://localhost:80/", url); !ok {
		t.Error()
	}

	url, err = mgr.resolveURL("zdb://127.0.0.1:80/", &api)

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "zdb://1.2.3.4:8080/", url); !ok {
		t.Error()
	}
}
