package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveInvalidSyntax(t *testing.T) {
	address := "1234"

	if ok := assert.Equal(t, address, Resolve(address)); !ok {
		t.Error()
	}

	address = ":1234"
	if ok := assert.Equal(t, address, Resolve(address)); !ok {
		t.Error()
	}
}

func TestResolveAny(t *testing.T) {
	src, _ := getSource("8080")
	socat.rules = map[int]rule{
		src.port: rule{
			source: src,
			port:   80,
			ip:     "1.2.3.4",
		},
	}

	if ok := assert.Equal(t, "1.2.3.4:80", Resolve("127.0.0.1:8080")); !ok {
		t.Error()
	}
}

func TestResolveNetwork(t *testing.T) {
	src, _ := getSource("127.0.0.0/8:8080")
	socat.rules = map[int]rule{
		src.port: rule{
			source: src,
			port:   80,
			ip:     "1.2.3.4",
		},
	}

	if ok := assert.Equal(t, "1.2.3.4:80", Resolve("127.0.0.1:8080")); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "128.0.0.1:8080", Resolve("128.0.0.1:8080")); !ok {
		t.Error()
	}
}

func TestResolveInf(t *testing.T) {
	src, _ := getSource("lo:8080")
	socat.rules = map[int]rule{
		src.port: rule{
			source: src,
			port:   80,
			ip:     "1.2.3.4",
		},
	}

	if ok := assert.Equal(t, "1.2.3.4:80", Resolve("127.0.0.1:8080")); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "192.168.1.1:8080", Resolve("192.168.1.1:8080")); !ok {
		t.Error()
	}
}

func TestResolveInfParital(t *testing.T) {
	src, _ := getSource("l*:8080")
	socat.rules = map[int]rule{
		src.port: rule{
			source: src,
			port:   80,
			ip:     "1.2.3.4",
		},
	}

	if ok := assert.Equal(t, "1.2.3.4:80", Resolve("127.0.0.1:8080")); !ok {
		t.Error()
	}

	if ok := assert.Equal(t, "192.168.1.1:8080", Resolve("192.168.1.1:8080")); !ok {
		t.Error()
	}
}

func TestResolveURL(t *testing.T) {
	src, _ := getSource(":80")
	socat.rules = map[int]rule{
		src.port: rule{
			source: src,
			port:   8080,
			ip:     "1.2.3.4",
		},
	}

	url, err := ResolveURL("http://127.0.0.1/")

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "http://1.2.3.4:8080/", url); !ok {
		t.Error()
	}

	url, err = ResolveURL("http://127.0.0.1:9000/")

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "http://127.0.0.1:9000/", url); !ok {
		t.Error()
	}

	url, err = ResolveURL("http://localhost/")

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "http://localhost:80/", url); !ok {
		t.Error()
	}

	url, err = ResolveURL("zdb://127.0.0.1:80/")

	if ok := assert.NoError(t, err); !ok {
		t.Fatal()
	}

	if ok := assert.Equal(t, "zdb://1.2.3.4:8080/", url); !ok {
		t.Error()
	}
}
