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

func TestCmdParsing(t *testing.T) {
	cmdline := parseKerenlOptions("zerotier=mynetwork")
	validateKeyValue(cmdline, "zerotier", "mynetwork", t)

	cmdline = parseKerenlOptions(`something   zerotier="my network"  rgergerger`)
	validateKeyValue(cmdline, "zerotier", "my network", t)
	if !cmdline.Is("something") {
		t.Error("`something` is not set")
	}
}

func TestCmdCmdline(t *testing.T) {
	cmdline := parseKerenlOptions(`something   zerotier="my network"  development`)
	validateKeyValue(cmdline, "zerotier", "my network", t)

	cl := cmdline.Cmdline()

	parsed := parseKerenlOptions(string(cl))

	if !parsed.Is("something") {
		t.Error("`something` is not set")
	}

	if !parsed.Is("development") {
		t.Error("`development` is not set")
	}

	if values, ok := parsed.Get("zerotier"); ok {
		if len(values) != 1 {
			t.Error("zerotier values should be of length 1")
		}
		if values[0] != "my network" {
			t.Error("zerotier value is wrong")
		}
	} else {
		t.Error("`zerotier` is not set")
	}
}
