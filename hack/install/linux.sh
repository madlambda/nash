#!/bin/bash

set -o errexit
set -o nounset

VERSION="v0.5"
echo "installing nash: "$VERSION" at NASHPATH: "$NASHPATH

cd $NASHPATH
tarfile="nash-$VERSION-linux-amd64.tar.gz"
wget https://github.com/NeowayLabs/nash/releases/download/$VERSION/$tarfile
tar xvfz $tarfile
rm -f $tarfile
