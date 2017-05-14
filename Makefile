all: build test install

build:
	cd cmd/nash && make -e build
	cd cmd/nashfmt && make -e build

deps:
	go get -v -t golang.org/x/exp/ebnf

docsdeps:
	go get github.com/katcipis/mdtoc

docs: docsdeps
	mdtoc -w ./README.md
	mdtoc -w ./docs/interactive.md
	mdtoc -w ./docs/reference.md

test: deps build
	GO15VENDOREXPERIMENT=1 ./hack/check.sh

install:
	cd cmd/nash && make -e install
	cd cmd/nashfmt && make -e install
	@echo "Nash installed on $(GOPATH)/bin/nash"

update-vendor:
	cd cmd/nash && nash ./vendor.sh

release: clean
	./hack/releaser.sh $(version)

clean:
	rm -f cmd/nash/nash
	rm -f cmd/nashfmt/nashfmt
	rm -rf dist
