#!/bin/bash

set -o errexit
set -o nounset

if [ -z "$NASHROOT" ]; then
        NASHROOT=$HOME/nashroot
fi

VERSION="v0.6"

if [ $# -eq 1 ]; then
        VERSION=$1
fi

echo "installing nash: "$VERSION" at NASHROOT: "$NASHROOT

mkdir -p $NASHROOT
cd $NASHROOT
tarfile="nash-$VERSION-linux-amd64.tar.gz"
wget https://github.com/NeowayLabs/nash/releases/download/$VERSION/$tarfile
tar xvfz $tarfile
rm -f $tarfile
