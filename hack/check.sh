#!/bin/bash

set -e

go test -race -coverprofile=coverage.txt ./...

echo "running stdlib and stdbin tests"
tests=$(find ./stdlib ./stdbin -name "*_test.sh")

for t in ${tests[*]}
do
    echo
    echo "running test: "$t
    ./cmd/nash/nash $t
    echo "success"
done
