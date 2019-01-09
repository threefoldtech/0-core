out = modules
modules = $(shell find . -maxdepth 1 -type d -not -name .)

all: $(modules)

.PHONY: output clean

output:
	mkdir -p $(out)

%: %/plugin output
	go build --buildmode=plugin -o $(out)/$(shell dirname $<).so  github.com/threefoldtech/0-core/apps/plugins/$<
