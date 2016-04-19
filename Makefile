all: build

build:
	cd cmd/cnt && make build

test: build
	./hack/check.sh
