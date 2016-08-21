all: build install

VERSION=$(shell git rev-parse --abbrev-ref HEAD)
BUILDARGS = -ldflags "-linkmode external -extldflags -static -X main.VersionString=$(VERSION)" -v

build:
	GO15VENDOREXPERIMENT=1 go build $(BUILDARGS)
install:
	GO15VENDOREXPERIMENT=1 go install $(BUILDARGS)
