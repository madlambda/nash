#!/usr/bin/env nash
fn vendor() {
        IFS = ()
	rm -rf vendor
	mkdir -p vendor/bin vendor/src vendor/pkg
	GOPATH <= pwd | xargs echo -n
	GOPATH = $GOPATH+"/vendor"
	setenv GOPATH
	GOBIN = $GOPATH+"/bin"
	setenv GOBIN
	go get -v .
	IFS = ("\n")
	paths <= ls vendor/src
	for path in $paths {
		mv "vendor/src/"+$path vendor/
	}
	rm -rf vendor/src vendor/bin vendor/pkg
	# because nash library is a dependency of cmd/nash
	# we need to remove it at end
	rm -rf vendor/github.com/NeowayLabs
}

vendor()
