all: build

build:
	cd cmd/cnt && make build

test: build
	GO15VENDOREXPERIMENT=1 ./hack/check.sh
