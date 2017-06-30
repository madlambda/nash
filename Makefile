ifndef version
version=$(shell git rev-list -1 HEAD)
endif

buildargs = -ldflags "-X main.VersionString=$(version)" -v

all: build test install

build:
	GO15VENDOREXPERIMENT=1 go build $(buildargs) -o ./cmd/nash/nash ./cmd/nash
	GO15VENDOREXPERIMENT=1 go build $(buildargs) -o ./cmd/nashfmt/nashfmt ./cmd/nashfmt

install:
ifndef NASHPATH
$(error NASHPATH must be set in order to install and use nash)
endif
	@echo "TODO"

deps:
	go get -v -t golang.org/x/exp/ebnf

docsdeps:
	go get github.com/katcipis/mdtoc

docs: docsdeps
	mdtoc -w ./README.md
	mdtoc -w ./docs/interactive.md
	mdtoc -w ./docs/reference.md
	mdtoc -w ./docs/stdlib/fmt.md

test: deps build
	GO15VENDOREXPERIMENT=1 ./hack/check.sh

update-vendor:
	cd cmd/nash && nash ./vendor.sh

release: clean
	./hack/releaser.sh $(version)

clean:
	rm -f cmd/nash/nash
	rm -f cmd/nashfmt/nashfmt
	rm -rf dist
