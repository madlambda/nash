#!/usr/bin/env nash

if len($ARGS) != "2" {
	print("usage: %s <version>\n\n", $ARGS[0])
	exit("1")
}

var version = $ARGS[1]
var supported_os = ("linux" "darwin" "windows")
var supported_arch = ("amd64")

# Guarantee passing tests at least on the host arch/os
make test
mkdir -p dist

fn prepare_execs(distfiles, os) {
	if $os == "windows" {
		var newfiles = ()
		
		for distfile in $distfiles {
			var src = $distfile[0]
			var dst = $distfile[1]
			var newsrc = $src+".exe"
			var newdst = $dst+".exe"
		
			newfiles <= append($newfiles, ($newsrc $newdst))
		}
		
		return $newfiles
	}
	if $os == "linux" {
		for distfile in $distfiles {
			strip $distfile[0]
		}
	}

	return $distfiles
}

for os in $supported_os {
	for arch in $supported_arch {
		setenv GOOS = $os
		setenv GOARCH = $arch

		if $os == "linux" {
			setenv CGO_ENABLED = "1"
		} else {
			setenv CGO_ENABLED = "0"
		}

		echo "building OS: "+$GOOS+" ARCH : "+$GOARCH
		make build "version="+$version

		var pkgdir <= mktemp -d

		var bindir = $pkgdir+"/bin"
		var stdlibdir = $pkgdir+"/stdlib"

		mkdir -p $bindir
		mkdir -p $stdlibdir

		var nash_src = "./cmd/nash/nash"
		var nash_dst = $bindir+"/nash"
		var nashfmt_src = "./cmd/nashfmt/nashfmt"
		var nashfmt_dst = $bindir+"/nashfmt"
		var execfiles = (
			($nash_src $nash_dst)
			($nashfmt_src $nashfmt_dst)
		)

		var execfiles <= prepare_execs($execfiles, $os)

		# TODO: Improve with glob, right now have only two packages =)
		var distfiles <= append($execfiles, ("./stdlib/io.sh" $stdlibdir))

		distfiles <= append($distfiles, ("./stdlib/map.sh" $stdlibdir))

		for distfile in $distfiles {
			var src = $distfile[0]
			var dst = $distfile[1]

			cp -pr $src $dst
		}

		var projectdir <= pwd
		var distar <= format("%s/dist/nash-%s-%s-%s.tar.gz", $projectdir, $version, $os, $arch)

		chdir($pkgdir)

		var pkgraw <= ls
		var pkgfiles <= split($pkgraw, "\n")

		tar cvfz $distar $pkgfiles

		chdir($projectdir)
	}
}
