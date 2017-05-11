#!/usr/bin/env nash

if len($ARGS) != "2" {
	print("usage: %s <version>\n\n", $ARGS[0])
	exit("1")
}

version      = $ARGS[1]
supported_os = ("linux" "darwin" "windows")

# Guarantee passing tests at least on the host arch/os
make test

setenv CGO_ENABLED = "0"

mkdir -p dist

for os in $supported_os {
	setenv GOOS = $os
	setenv GOARCH = "amd64"

	echo "building OS: "+$GOOS+" ARCH : "+$GOARCH
	make build "version="+$version

	binpath <= format("dist/nash-%s-%s-amd64", $version, $os)

	if $os != "windows" {
		binpath = $binpath+".exe"
	}

	cp "cmd/nash/nash" $binpath
}
