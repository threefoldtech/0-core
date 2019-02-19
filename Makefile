OUTPUT = bin
VERSION = base/version.go

branch = $(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)
revision = $(shell git rev-parse HEAD)
dirty = $(shell test -n "`git diff --shortstat 2> /dev/null | tail -n1`" && echo "*")
base = github.com/threefoldtech/0-core/base
ldflags0 = '-w -s -X $(base).Branch=$(branch) -X $(base).Revision=$(revision) -X $(base).Dirty=$(dirty)'
ldflagsX = '-w -s -X $(base).Branch=$(branch) -X $(base).Revision=$(revision) -X $(base).Dirty=$(dirty) -extldflags "-static"'

all: core0 coreX corectl redis-proxy

core0: $(OUTPUT)
	cd apps/core0 && go build -ldflags $(ldflags0) -o ../../$(OUTPUT)/$@

coreX: $(OUTPUT)
	cd apps/coreX && GOOS=linux go build -ldflags $(ldflagsX) -o ../../$(OUTPUT)/$@

corectl: $(OUTPUT)
	cd apps/corectl && go build -ldflags $(ldflags0) -o ../../$(OUTPUT)/$@

redis-proxy: $(OUTPUT)
	cd apps/redis-proxy && go build -ldflags $(ldflags0) -o ../../$(OUTPUT)/$@

plugins:
	cd apps/plugins && make

$(OUTPUT):
	mkdir -p $(OUTPUT)

.PHONY: $(OUTPUT) core0 coreX corectl redis-proxy
