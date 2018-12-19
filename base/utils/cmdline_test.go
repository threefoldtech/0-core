package utils

import (
	"fmt"
	"testing"
)

func validateKeyValue(mapping KernelOptions, key string, value string, t *testing.T) {
	if val, ok := mapping[key]; ok {
		if val[0] != value {
			t.Fatal(fmt.Printf("Value %s != %s", val, value))
		}
	} else {
		t.Fatal(fmt.Printf("Key %s missing", key))
	}

}

func validateMultiValueForKey(opts KernelOptions, key string, values []string, t *testing.T) {
	if vals, ok := opts[key]; ok {
		if len(vals) != len(values) {
			t.Fatalf("no multiple values for %s", key)
		}
		for i, v := range vals {
			t.Logf("got value %+v", v)
			if v != values[i] {
				t.Fatalf("%s index %d is not eq to %s", values[i], i, v)
			}
		}
	}
	t.Logf("%+v", opts)
}

func TestCmdParsing(t *testing.T) {
	cmdline := parseKernelOptions("zerotier=mynetwork")
	validateKeyValue(cmdline, "zerotier", "mynetwork", t)

	cmdline = parseKernelOptions(`something   zerotier="my network"  rgergerger`)
	validateKeyValue(cmdline, "zerotier", "my network", t)

	cmdline = parseKernelOptions("noautonic=enp2s0f0 noautonic=enp2s0f1")
	vals := []string{"enp2s0f0", "enp2s0f1"}
	validateMultiValueForKey(cmdline, "noautonic", vals, t)
}
