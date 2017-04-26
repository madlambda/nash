all: build install

ifndef VERSION
VERSION=$(shell git rev-list -1 HEAD)
endif

BUILDARGS = -installsuffix netgo -ldflags "-linkmode external -extldflags -static -X main.VersionString=$(VERSION)" -v
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	BUILDARGS = -ldflags "-linkmode external -X main.VersionString=$(VERSION)" -v
endif

build:
	@echo "building nash version: "$(VERSION)
	GO15VENDOREXPERIMENT=1 go build $(BUILDARGS)
install:
	GO15VENDOREXPERIMENT=1 go install $(BUILDARGS)
