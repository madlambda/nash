#!/usr/bin/env nash

if len($ARGS) != "2" {
	print("usage: %s <version>\n\n", $ARGS[0])
	exit("1")
}

version = $ARGS[1]
supported_os = ("linux" "darwin" "windows")

# Guarantee passing tests at least on the host arch/os
make test

setenv CGO_ENABLED="0"
mkdir -p dist

for os in $supported_os {
    setenv GOOS = $os
    setenv GOARCH = "amd64"
    echo "building OS: " + $GOOS + " ARCH : " + $GOARCH
    make build "version=" + $version
    if $os != "windows" {
        cp "cmd/nash/nash" "dist/nash-"+$os+"-amd64"
    } else {
        cp cmd/nash/nash.exe "dist/nash-"+$os+"-amd64.exe"
    }
}
