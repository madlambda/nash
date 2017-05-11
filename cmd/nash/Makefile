all: build install

ifndef version
version=$(shell git rev-list -1 HEAD)
endif

buildargs = -ldflags "-X main.VersionString=$(version)" -v

build:
	@echo "building nash version: "$(version)
	GO15VENDOREXPERIMENT=1 go build $(buildargs)

install:
	GO15VENDOREXPERIMENT=1 go install $(buildargs)
