package socat

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSource(t *testing.T) {

	for _, tc := range []struct {
		input    string
		expected source
	}{
		{
			"2200",
			source{ip: "0.0.0.0", port: 2200, protocols: defaultProtocols},
		},
		{
			"0.0.0.0:2200",
			source{ip: "0.0.0.0", port: 2200, protocols: defaultProtocols},
		},
		{
			"10.20.30.40:2200",
			source{ip: "10.20.30.40", port: 2200, protocols: defaultProtocols},
		},
		{
			"eth0:2200",
			source{ip: "eth0", port: 2200, protocols: defaultProtocols},
		},
		{
			"53|udp",
			source{ip: "0.0.0.0", port: 53, protocols: []string{"udp"}},
		},
		{
			"8000|tcp+udp",
			source{ip: "0.0.0.0", port: 8000, protocols: []string{"tcp", "udp"}},
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			src, err := getSource(tc.input)
			require.NoError(t, err)
			if ok := assert.Equal(t, tc.expected, src); !ok {
				t.Error()
			}
		})
	}

	for _, tc := range []string{"eth0:0", "eth0", "", "123:http"} {
		t.Run(tc, func(t *testing.T) {
			_, err := getSource(tc)
			require.Error(t, err)
		})
	}
}

func TestRule(t *testing.T) {
	source, err := getSource("80|tcp+udp")
	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	r := rule{
		source: source,
		port:   8080,
		ip:     "1.2.3.4",
	}

	rules := r.Rules()

	if ok := assert.Len(t, rules, 2); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(
		t,
		"ip daddr @host tcp dport 80 meta mark set 0 dnat to 1.2.3.4:8080",
		rules[0],
	); !ok {
		t.Error()
	}

	if ok := assert.Equal(
		t,
		"ip daddr @host udp dport 80 meta mark set 0 dnat to 1.2.3.4:8080",
		rules[1],
	); !ok {
		t.Error()
	}
}

func TestRuleString(t *testing.T) {

	var tests = []string{
		"80|tcp+udp",
		"80",
		"zt*:80|udp",
		"10.20.30.40:1001|tcp+udp",
		"10.20.0.0/16:1200",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			source, err := getSource(input)
			if ok := assert.NoError(t, err); !ok {
				t.Fatal()
			}

			if ok := assert.Equal(t, input, source.String()); !ok {
				t.Error()
			}
		})
	}

}

func TestRuleFromNFT(t *testing.T) {
	/*
		ip daddr @host iifname "zt*" tcp dport 1028 mark set 0x01000002 dnat to 172.18.0.3:6379
		ip daddr @host iifname "zt*" udp dport 1028 mark set 0x01000002 dnat to 172.18.0.3:6379
		ip daddr @host ip saddr 10.20.100.100 tcp dport 1029 mark set 0x01000002 dnat to 172.18.0.3:6379
		ip daddr @host ip saddr 192.168.0.0/16 udp dport 6378 mark set 0x01000002 dnat to 172.18.0.3:6379
	*/

	var tests = map[string]rule{
		"ip daddr @host iifname \"zt*\" tcp dport 1028 mark set 0x01000002 dnat to 172.18.0.3:6379": rule{
			ns:   0x01000002,
			ip:   "172.18.0.3",
			port: 6379,
			source: source{
				ip:        "zt*",
				port:      1028,
				protocols: []string{"tcp"},
			},
		},
		"ip daddr @host iifname \"zt*\" udp dport 1028 mark set 0x01000002 dnat to 172.18.0.3:6379": rule{
			ns:   0x01000002,
			ip:   "172.18.0.3",
			port: 6379,
			source: source{
				ip:        "zt*",
				port:      1028,
				protocols: []string{"udp"},
			},
		},
		"ip daddr @host ip saddr 10.20.100.100 tcp dport 1029 mark set 0x020000B2 dnat to 172.18.0.3:6379": rule{
			ns:   0x020000B2,
			ip:   "172.18.0.3",
			port: 6379,
			source: source{
				ip:        "10.20.100.100",
				port:      1029,
				protocols: []string{"tcp"},
			},
		},
		"ip daddr @host ip saddr 192.168.0.0/16 udp dport 6378 mark set 0x010000a3 dnat to 172.18.0.3:6379": rule{
			ns:   0x010000a3,
			ip:   "172.18.0.3",
			port: 6379,
			source: source{
				ip:        "192.168.0.0/16",
				port:      6378,
				protocols: []string{"udp"},
			},
		},
	}

	for ruleStr, expected := range tests {
		t.Run(ruleStr, func(t *testing.T) {
			parsed, err := getRuleFromNFTRule(ruleStr)
			require.NoError(t, err)
			require.Equal(t, expected, parsed)
		})
	}
}
