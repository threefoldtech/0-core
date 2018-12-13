out = output
modules = $(shell find . -maxdepth 1 -type d -not -name .)

.PHONY: output clean
output:
	mkdir -p $(out)

all: $(modules)

%: %/plugin output
	go build --buildmode=plugin -o $(out)/$(shell dirname $<).so  github.com/threefoldtech/0-core/apps/plugins/$<
