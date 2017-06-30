ifndef version
version=$(shell git rev-list -1 HEAD)
endif

buildargs = -ldflags "-X main.VersionString=$(version)" -v

all: build test install

build:
	go build $(buildargs) -o ./cmd/nash/nash ./cmd/nash
	go build $(buildargs) -o ./cmd/nashfmt/nashfmt ./cmd/nashfmt


guard-%:
	@ if [ "${${*}}" = "" ]; then \
                echo "'$*' must be set in order to install and use nash"; \
                exit 1; \
        fi

install: build guard-NASHPATH
	@echo
	@echo "Installing nash at: "$(NASHPATH)
	mkdir -p $(NASHPATH)/bin
	mkdir -p $(NASHPATH)/lib
	rm -f $(NASHPATH)/bin/nash
	rm -f $(NASHPATH)/bin/nashfmt
	cp -p ./cmd/nash/nash $(NASHPATH)/bin
	cp -p ./cmd/nashfmt/nashfmt $(NASHPATH)/bin
	rm -rf $(NASHPATH)/stdlib
	cp -pr ./stdlib $(NASHPATH)/stdlib

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
	./hack/check.sh

update-vendor:
	cd cmd/nash && nash ./vendor.sh

release: clean
	./hack/releaser.sh $(version)

clean:
	rm -f cmd/nash/nash
	rm -f cmd/nashfmt/nashfmt
	rm -rf dist
