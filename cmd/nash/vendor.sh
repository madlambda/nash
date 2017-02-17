#!/usr/bin/env nash

fn vendor() {
        cwdir <= pwd | xargs echo -n
        vendordir = $cwdir + "/vendor"
        rm -rf $vendordir

        bindir = $vendordir + "/bin"
        srcdir = $vendordir + "/src"
        pkgdir = $vendordir + "/pkg"
        mkdir -p $bindir $srcdir $pkgdir

        setenv GOPATH = $vendordir
        setenv GOBIN = $vendordir

        go get -v .

        rawpaths <= ls $srcdir
        paths <= split($paths, "\n")
        for path in $paths {
                mv $srcdir + $path $vendor
        }
        rm -rf $bindir $srcdir $pkgdir

        # because nash library is a dependency of cmd/nash
        # we need to remove it at end
        rm -rf vendor/github.com/NeowayLabs
}

vendor()
