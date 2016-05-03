#!/usr/bin/env nash
# Build the stali rootfs
# Requires: build-essential mawk

-rm -rf toolchain
-rm -rf src
-rm -rf rootfs-x86_64/

mkdir rootfs-x86_64

git clone --depth=1 http://git.sta.li/toolchain
git clone --depth=1 http://git.sta.li/src

STALI_SRC=$PWD + "/src"

mv src/config.mk src/config.mk.orig
cp config.mk src/config.mk

cd src

make
make install

cd ..
tar cvf rootfs-x86_64.tar rootfs-x86_64
bzip2 rootfs-x86_64.tar

echo "Stali image generated."
