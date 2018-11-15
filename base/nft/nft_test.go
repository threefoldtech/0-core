package nft

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	nft := Nft{
		"nat": Table{
			Family: FamilyIP,
			Chains: Chains{
				"post": Chain{
					Rules: []Rule{
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

func TestFindRules(t *testing.T) {
	config := `{"nftables": [{"table": {"family": "ip", "name": "nat", "handle": 0}}, {"set": {"family": "ip", "elem": ["10.20.1.1", "172.18.0.1", "172.19.0.1"], "name": "host", "table": "nat", "handle": 0, "type": "ipv4_addr"}}, {"chain": {"hook": "prerouting", "family": "ip", "table": "nat", "prio": 0, "name": "pre", "type": "nat", "handle": 1, "policy": "accept"}}, {"rule": {"chain": "pre", "family": "ip", "table": "nat", "handle": 7, "expr": [{"match": {"left": {"payload": {"field": "daddr", "name": "ip"}}, "right": "@host"}}, {"match": {"left": {"meta": "iifname"}, "right": "eth0"}}, {"match": {"left": {"payload": {"field": "dport", "name": "tcp"}}, "right": 8000}}, {"dnat": {"addr": "172.18.0.2", "port": 7999}}]}}, {"chain": {"hook": "postrouting", "family": "ip", "table": "nat", "prio": 0, "name": "post", "type": "nat", "handle": 2, "policy": "accept"}}, {"rule": {"chain": "post", "family": "ip", "table": "nat", "handle": 4, "expr": [{"match": {"left": {"payload": {"field": "saddr", "name": "ip"}}, "right": {"prefix": {"addr": "172.18.0.0", "len": 16}}}}, {"masquerade": null}]}}, {"rule": {"chain": "post", "family": "ip", "table": "nat", "handle": 5, "expr": [{"match": {"left": {"payload": {"field": "saddr", "name": "ip"}}, "right": {"prefix": {"addr": "172.19.0.0", "len": 16}}}}, {"masquerade": null}]}}, {"table": {"family": "inet", "name": "filter", "handle": 0}}, {"chain": {"hook": "prerouting", "family": "inet", "table": "filter", "prio": 0, "name": "pre", "type": "filter", "handle": 1, "policy": "accept"}}, {"chain": {"hook": "forward", "family": "inet", "table": "filter", "prio": 0, "name": "forward", "type": "filter", "handle": 2, "policy": "accept"}}, {"chain": {"hook": "output", "family": "inet", "table": "filter", "prio": 0, "name": "output", "type": "filter", "handle": 3, "policy": "accept"}}, {"chain": {"hook": "input", "family": "inet", "table": "filter", "prio": 0, "name": "input", "type": "filter", "handle": 4, "policy": "drop"}}, {"rule": {"chain": "input", "family": "inet", "table": "filter", "handle": 5, "expr": [{"match": {"left": {"ct": {"key": "state"}}, "right": {"set": ["established", "related"]}}}, {"accept": null}]}}, {"rule": {"chain": "input", "family": "inet", "table": "filter", "handle": 6, "expr": [{"match": {"left": {"meta": "iifname"}, "right": "lo"}}, {"accept": null}]}}, {"rule": {"chain": "input", "family": "inet", "table": "filter", "handle": 7, "expr": [{"match": {"left": {"meta": "iifname"}, "right": "vxbackend"}}, {"accept": null}]}}, {"rule": {"chain": "input", "family": "inet", "table": "filter", "handle": 8, "expr": [{"match": {"left": {"payload": {"field": "protocol", "name": "ip"}}, "right": 1}}, {"accept": null}]}}, {"rule": {"chain": "input", "family": "inet", "table": "filter", "handle": 9, "expr": [{"match": {"left": {"meta": "iif"}, "right": "core0"}}, {"match": {"left": {"payload": {"field": "dport", "name": "udp"}}, "right": {"set": [53, 67, 68]}}}, {"accept": null}]}}, {"rule": {"chain": "input", "family": "inet", "table": "filter", "handle": 10, "expr": [{"match": {"left": {"payload": {"field": "dport", "name": "tcp"}}, "right": 22}}, {"accept": null}]}}, {"rule": {"chain": "input", "family": "inet", "table": "filter", "handle": 11, "expr": [{"match": {"left": {"payload": {"field": "dport", "name": "tcp"}}, "right": 6379}}, {"accept": null}]}}, {"rule": {"chain": "input", "family": "inet", "table": "filter", "handle": 12, "expr": [{"match": {"left": {"meta": "iif"}, "right": "kvm0"}}, {"match": {"left": {"payload": {"field": "dport", "name": "udp"}}, "right": {"set": [53, 67, 68]}}}, {"accept": null}]}}, {"rule": {"chain": "input", "family": "inet", "table": "filter", "handle": 13, "expr": [{"match": {"left": {"payload": {"field": "dport", "name": "tcp"}}, "right": 5900}}, {"accept": null}]}}]}`

	ruleset, err := Parse(config)

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	sub := Nft{
		"nat": Table{
			Family: FamilyIP,
			Chains: Chains{
				"pre": Chain{
					Rules: []Rule{
						{Body: "ip daddr @host iifname \"eth0\" tcp dport 8000 dnat to 172.18.0.2:7999"},
					},
				},
			},
		},
		"filter": Table{
			Family: FamilyINET,
			Chains: Chains{
				"input": Chain{
					Rules: []Rule{
						{Body: "tcp dport 5900 accept"},
					},
				},
			},
		},
	}

	sub, err = findRules(ruleset, sub)
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, 7, sub["nat"].Chains["pre"].Rules[0].Handle); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, 13, sub["filter"].Chains["input"].Rules[0].Handle); !ok {
		t.Error()
	}
}
