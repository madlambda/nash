ifndef version
version=$(shell git rev-list -1 HEAD)
endif

buildargs = -ldflags "-X main.VersionString=$(version)" -v

all: build test install

build:
	cd cmd/nash && go build $(buildargs) 
	cd cmd/nashfmt && go build $(buildargs) 
	cd stdbin/mkdir && go build $(buildargs)
	cd stdbin/pwd && go build $(buildargs)
	cd stdbin/write && go build $(buildargs)
	cd stdbin/strings && go build $(buildargs)

NASHPATH?=$(HOME)/nash
NASHROOT?=$(HOME)/nashroot

# FIXME: binaries install do not work on windows this way (the .exe)
install: build
	@echo
	@echo "Installing nash at: "$(NASHROOT)
	mkdir -p $(NASHROOT)/bin
	rm -f $(NASHROOT)/bin/nash
	rm -f $(NASHROOT)/bin/nashfmt
	cp -p ./cmd/nash/nash $(NASHROOT)/bin
	cp -p ./cmd/nashfmt/nashfmt $(NASHROOT)/bin
	rm -rf $(NASHROOT)/stdlib
	cp -pr ./stdlib $(NASHROOT)/stdlib
	cp -pr ./stdbin/mkdir/mkdir $(NASHROOT)/bin/mkdir
	cp -pr ./stdbin/pwd/pwd $(NASHROOT)/bin/pwd
	cp -pr ./stdbin/write/write $(NASHROOT)/bin/write
	cp -pr ./stdbin/strings/strings $(NASHROOT)/bin/strings

docsdeps:
	go get github.com/madlambda/mdtoc/cmd/mdtoc

docs: docsdeps
	mdtoc -w ./README.md
	mdtoc -w ./docs/interactive.md
	mdtoc -w ./docs/reference.md
	mdtoc -w ./docs/stdlib/fmt.md

test: build
	./hack/check.sh

update-vendor:
	cd cmd/nash && nash ./vendor.sh

release: clean
	./hack/releaser.sh $(version)

coverage-html: test
	go tool cover -html=coverage.txt -o coverage.html
	@echo "coverage file: coverage.html"

coverage-show: coverage-html
	xdg-open coverage.html

clean:
	rm -f cmd/nash/nash
	rm -f cmd/nashfmt/nashfmt
	rm -rf dist
