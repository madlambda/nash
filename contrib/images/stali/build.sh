#!/usr/bin/env nash
# Build the stali rootfs
# Requires: build-essential mawk

imagename = "stali-x86_64"

fn cleanup() {
    -rm -rf toolchain
    -rm -rf src
    -rm -rf rootfs-x86_64/
}

fn download() {
    git clone --depth=1 http://git.sta.li/toolchain
    git clone --depth=1 http://git.sta.li/src
}

fn buildStali() {
    cleanup()
    download()

    mkdir $imagename
    STALI_SRC = $PWD + "/src"

    mv src/config.mk src/config.mk.orig
    cp config.mk src/config.mk

    cd src

    make
    make install

    cd ..
    tar cvf $imagename+".tar" $imagename
    bzip2 $imagename+".tar"

    echo "Stali image generated: " + $imagename + ".tar.bz2"

    out = $PWD + "/" + $imagename + ".tar.bz2"

    return $out
}

buildStali()
