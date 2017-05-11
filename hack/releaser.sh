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

fn copy_exec(src, dst) {
	if $os == "windows" {
		src = $src+".exe"
		dst = $dst+".exe"
	}

	cp $src $dst
}

for os in $supported_os {
	setenv GOOS = $os
	setenv GOARCH = "amd64"

	echo "building OS: "+$GOOS+" ARCH : "+$GOARCH
	make build "version="+$version

	source_nash = "cmd/nash/nash"

	target_nash <= format("dist/nash-%s-%s-amd64", $version, $os)

	copy_exec($source_nash, $target_nash)

	source_nashfmt = "cmd/nashfmt/nashfmt"

	target_nashfmt <= format("dist/nashfmt-%s-%s-amd64", $version, $os)

	copy_exec($source_nashfmt, $target_nashfmt)
}
