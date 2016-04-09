all: build

build:
	cd cmd/cnt && make build

test:
	./hack/check.sh
