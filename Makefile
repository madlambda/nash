all: build

build:
	cd cmd/nash && make build

test: build
	GO15VENDOREXPERIMENT=1 ./hack/check.sh
