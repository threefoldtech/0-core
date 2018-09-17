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
			source{"0.0.0.0", 2200},
		},
		{
			"0.0.0.0:2200",
			source{"0.0.0.0", 2200},
		},
		{
			"10.20.30.40:2200",
			source{"10.20.30.40", 2200},
		},
		{
			"eth0:2200",
			source{"eth0", 2200},
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			src, err := getSource(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, src)
		})
	}

	for _, tc := range []string{"eth0:0", "eth0", ""} {
		t.Run(tc, func(t *testing.T) {
			_, err := getSource(tc)
			assert.Error(t, err)
		})
	}
}

func TestCompareSource(t *testing.T) {
	onlyPort := source{
		ip:   "",
		port: 80,
	}
	interfacePort := source{
		ip:   "zt*",
		port: 80,
	}
	rules := map[source]rule{
		onlyPort: rule{},
	}

	_, exists := rules[onlyPort]
	require.True(t, exists)

	_, exists = rules[interfacePort]
	require.False(t, exists)
}
