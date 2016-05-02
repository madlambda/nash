#!/usr/bin/env nash
# Build the stali rootfs
# Requires: build-essential mawk

-rm -rf toolchain
-rm -rf src
mkdir rootfs-x86_64

git clone http://git.sta.li/toolchain
git clone http://git.sta.li/src

STALI_SRC=$PWD + "/src"

cd src
mv config.mk config.mk.orig
cat config.mk.orig | sed 's/DESTDIR\=.*/DESTDIR=$(PWD)/rootfs-x86_64/g' > config.mk
make
make install

cd ..
tar cvf rootfs-x86_64.tar rootfs-x86_64
bzip2 rootfs-x86_64.tar

echo "Stali image generated."
