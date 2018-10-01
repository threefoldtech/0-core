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
			assert.Equal(t, tc.expected, src)
		})
	}

	for _, tc := range []string{"eth0:0", "eth0", "", "123:http"} {
		t.Run(tc, func(t *testing.T) {
			_, err := getSource(tc)
			assert.Error(t, err)
		})
	}
}
