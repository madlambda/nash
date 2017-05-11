all: build install

ifndef VERSION
VERSION=$(shell git rev-list -1 HEAD)
endif

BUILDARGS = -ldflags "-X main.VersionString=$(VERSION)" -v
UNAME_S := $(shell uname -s)

build:
	@echo "building nash version: "$(VERSION)
	GO15VENDOREXPERIMENT=1 go build $(BUILDARGS)

install:
	GO15VENDOREXPERIMENT=1 go install $(BUILDARGS)
