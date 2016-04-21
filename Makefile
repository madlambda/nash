all: build

build:
	cd cmd/nash && make build

deps:
	go get -v -t golang.org/x/exp/ebnf

test: deps build
	GO15VENDOREXPERIMENT=1 ./hack/check.sh
