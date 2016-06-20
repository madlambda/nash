all: build install

VERSION=$(shell git rev-parse --abbrev-ref HEAD)

build:
	GO15VENDOREXPERIMENT=1 go build -v -ldflags "-X main.VersionString=$(VERSION)"

install:
	GO15VENDOREXPERIMENT=1 go install -ldflags "-X main.VersionString=$(VERSION)"
