#!/usr/bin/env nash

-rm -rf rootfs

rfork upmis {
    mount -t proc proc /proc
    mkdir -p rootfs
    mount -t tmpfs -o size=1G tmpfs rootfs

    cd rootfs

    wget "https://busybox.net/downloads/binaries/latest/busybox-x86_64" -O busybox
    chmod +x busybox

    mkdir bin

    ./busybox --install ./bin

    mkdir -p proc
    mkdir -p dev
    mount -t proc proc proc
    mount -t tmpfs tmpfs dev

    cp ../nash .
    chroot . /bin/sh
}
