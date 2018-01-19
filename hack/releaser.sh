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
			src = $distfile[0]
			dst = $distfile[1]
			newsrc = $src + ".exe"
			newdst = $dst + ".exe"
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

		echo "building OS: "+$GOOS+" ARCH : "+$GOARCH
		make build "version="+$version

		pkgdir    <= mktemp -d
		bindir = $pkgdir + "/bin"
		stdlibdir = $pkgdir + "/stdlib"
		mkdir -p $bindir
		mkdir -p $stdlibdir

		nash_src = "./cmd/nash/nash"
		nash_dst = $bindir + "/nash"
		nashfmt_src = "./cmd/nashfmt/nashfmt"
		nashfmt_dst = $bindir + "/nashfmt"

		execfiles = ( ($nash_src $nash_dst) ($nashfmt_src $nashfmt_dst) )
		execfiles <= prepare_execs($execfiles, $os)

		# TODO: Improve with glob, right now have only two packages =)
		distfiles <= append($execfiles, ("./stdlib/io.sh" $stdlibdir))
		distfiles <= append($distfiles, ("./stdlib/map.sh" $stdlibdir))

		for distfile in $distfiles {
			src = $distfile[0]
			dst = $distfile[1]
			cp -pr $src $dst
		}

		projectdir <= pwd
		distar  <= format("%s/dist/nash-%s-%s-%s.tar.gz", $projectdir, $version, $os, $arch)

		chdir($pkgdir)
		pkgraw <= ls
		pkgfiles <= split($pkgraw, "\n")
		tar cvfz $distar $pkgfiles
		chdir($projectdir)
	}
}
