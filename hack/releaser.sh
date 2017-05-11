#!/usr/bin/env nash

if len($ARGS) != "2" {
	print("usage: %s <version>\n\n", $ARGS[0])
	exit("1")
}

version        = $ARGS[1]
supported_os   = ("linux" "darwin" "windows")
supported_arch = ("amd64")

# Guarantee passing tests at least on the host arch/os
make test

setenv CGO_ENABLED = "0"

mkdir -p dist

fn prepare_execs(distfiles, os) {
	if $os == "windows" {
		newfiles = ()
		
		for distfile in $distfiles {
			file = $distfile+".exe"
		
			newfiles <= append($newfiles, $file)
		}
		
		return $newfiles
	}
	if $os == "linux" {
		for distfile in $distfiles {
			strip $distfile
		}
	}

	return $distfiles
}

for os in $supported_os {
	for arch in $supported_arch {
		setenv GOOS = $os
		setenv GOARCH = $arch

		echo "building OS: "+$GOOS+" ARCH : "+$GOARCH
		make build "version="+$version

		nash      = "cmd/nash/nash"
		nashfmt   = "cmd/nashfmt/nashfmt"
		execfiles = ($nash $nashfmt)

		distfiles <= prepare_execs($execfiles, $os)
		distar    <= format("dist/nash-%s-%s-%s.tar.gz", $version, $os, $arch)

		tar cvfz $distar $distfiles
	}
}
