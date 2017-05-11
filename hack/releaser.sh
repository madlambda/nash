#!/usr/bin/env nash

supported_os = ("linux" "darwin" "windows")

# Guarantee passing tests at least on the host arch/os
make test

setenv CGO_ENABLED="0"

for os in $supported_os {
    setenv GOOS = $os
    setenv GOARCH = "amd64"
    echo "building OS: " + $GOOS + " ARCH : " + $GOARCH
    make build
    mkdir -p dist
    if $os != "windows" {
        cp "cmd/nash/nash" "dist/nash-"+$os+"-amd64"
    } else {
        cp cmd/nash/nash.exe "dist/nash-"+$os+"-amd64.exe"
    }
}
