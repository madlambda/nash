#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

which wget >/dev/null   || { echo "wget not installed"; exit 1; } 
which tar >/dev/null    || { echo "tar not installed"; exit 1; }
which tr >/dev/null     || { echo "tr not found"; exit 1; }

NASHROOT=${NASHROOT:-$HOME/nashroot}
VERSION="v1.1"
ARCH="amd64"
OS="$(uname | tr '[:upper:]' '[:lower:]')"

if [ $# -eq 1 ]; then
        VERSION=$1
fi

echo "installing nash (${OS}): ${VERSION} at NASHROOT: ${NASHROOT}"

mkdir -p $NASHROOT
cd $NASHROOT
tarfile="nash-${VERSION}-${OS}-${ARCH}.tar.gz"
wget -c https://github.com/madlambda/nash/releases/download/$VERSION/$tarfile
tar xvfz $tarfile
rm -f $tarfile
