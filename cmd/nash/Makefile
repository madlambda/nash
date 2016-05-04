all: build

VERSION=$(shell git rev-parse --abbrev-ref HEAD)

build:
	GO15VENDOREXPERIMENT=1 go build -v -ldflags "-X main.VersionString=$(VERSION)"
