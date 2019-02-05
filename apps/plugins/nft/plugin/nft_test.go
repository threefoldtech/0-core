package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/0-core/apps/plugins/nft"
)

func TestMarshal(t *testing.T) {
	nft := nft.Nft{
		"nat": nft.Table{
			Family: nft.FamilyIP,
			Chains: nft.Chains{
				"post": nft.Chain{
					Rules: []nft.Rule{
						{Body: fmt.Sprintf("ip saddr %s masquerade", "1.1.1.1")},
					},
				},
			},
		},
	}

	data, err := nft.MarshalText()

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}
	expected := `table ip nat {
	chain post {
		ip saddr 1.1.1.1 masquerade
	}
}
`

	if ok := assert.Equal(t, expected, string(data)); !ok {
		t.Error()
	}
}
