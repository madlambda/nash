# nash

[![Build Status](https://travis-ci.org/tiago4orion/nash.svg?branch=master)](https://travis-ci.org/tiago4orion/nash) [![codecov.io](https://codecov.io/github/tiago4orion/nash/coverage.svg?branch=master)](https://codecov.io/github/tiago4orion/nash?branch=master)

Nash is a Linux system shell that attempts to be more safe and give
more power to user. It's safe in the sense of it's far more hard to
shoot yourself in the foot, (or in the head in case of bash). It gives
power to the user in the sense that you really can use `nash` to
script deploys taking advantage of linux namespaces, cgroups, unionfs,
etc, in an idiomatic way.

It's more safe for devops/sysadmins because it doesn't have the unsafe
features of all former shells and it aim to have a very simple, safe
and frozen syntax specification. Every shell feature considered
harmful was left behind, but some needed features are still missing.

Nash is inspired by Plan 9 `rc` shell, but with very different syntax
and purpose.

# Concept

Nowadays everyone agrees that a good deploy requires containers, but
why this kind of tools (docker, rkt, etc) and libraries (lxc,
libcontainer, etc) are so bloated and magical?

In the past, the UNIX sysadmin had the complete understanding of the
operating system and the software being deployed. All of the operating
system packages/libraries going to production and the required network
configurations in every machine was maintained by well-know
scripts. Today we know that this approach have lots of problems and
the container approach is a better alternative.

Before Linux namespace, BSD Jails, Solaris Zones, and so on, the
sysadmin had to fight the global view of the operating
system. There was only one root mount table, only one view of devices
and processes, and so on. It was a mess. How to scale multiple
versions of the same app, in the same machine, if the app write to a
fixed path in the filesystem? Or, how to avoid clash of port numbers
when scaling apps?

But when the container idea arrived in the Linux world, it was in a
completely different way. Instead of giving the power of namespaces to
programmers or ops guys, the concept was hidden inside black
boxes. I'm not saying the current solutions will never work, but the
way it works could be harmful for community. Docker uses mount
namespaces? If yes, why the global mount table is dirty after a docker
run? Why docker needs root if linux kernel supports user namespace?
What technologies are used for docker networking? Do you know how
docker implements the union fs (layered fs) approach? Aufs, Union fs
or Device Mapper? Are it using chroot in addition to containers? But
how volumes are handled? If you know the answer for some of the
questions above, you must know that the answer changes a lot depending
on the operating system and docker release.

Nasn is a way for you, that understand the rules of the game, to make
reliable deploy scripts using the good parts of the container
technology.

Nash is only a simple shell plus a keyword called `rfork`. Rfork try
to mimic what Plan9 `rfork` does for namespaces, but with linux
limitations in mind.

Basically, your script can download an Operating System rootfs,
copy/install your application inside, start the needed namespaces for
serving this rootfs via chroot and then start the app. Or, if your
application is statically compiled, create an empty directory, copy
your application/micro-service into it, start the needed namespaces
and run it.

# Show time!

Go ahead:

```sh
go get github.com/tiago4orion/nash/cmd/nash
# Make sure GOPATH/bin is in yout PATH
nash
λ> echo "hello world"
hello world
```

Make sure you have USER namespaces enabled in your kernel:

```sh
zgrep CONFIG_USER_NS /proc/config.gz
CONFIG_USER_NS=y
```

If it's not enabled you will need root privileges to execute every example below...

Creating a new process in a new USER namespace (u):

```sh
λ> rfork u {
     id
}
uid=0(root) gid=0(root) groups=0(root),65534
```
Yes, Linux supports creation of containers by unprivileged users. Tell
this to the customer success of your container-infrastructure-vendor.

The default UID mapping is: Current UID (getuid) => 0 (no
range support). I'll look into more options for this in the future.

Yes, you can create multiple nested user namespaces. But kernel limits
the number of nested user namespace clones to 32.

```sh
λ> rfork u {
    echo "inside first container"

    id

    rfork u {
        echo "inside second namespace..."

        id
    }
}
```

You can verify that other types of namespace still requires root
capabilities, see for PID namespaces (p).

```sh
λ> rfork p {
    id
}
ERROR: fork/exec ./nash: operation not permitted
```

The same happens for mount (m), ipc (i) and uts (s) if used without
user namespace (u) flag.

The `c` flag stands for "container" and is an alias for upmnis (all
types of namespaces).  If you want a shell inside the container:

```sh
λ> rfork c {
    bash
}
[root@stay-away nash]# id
uid=0(root) gid=0(root) groups=0(root),65534
[root@stay-away nash]# mount -t proc proc /proc
[root@stay-away nash]#
[root@stay-away nash]# ps aux
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.0  0.0  34648  2748 pts/4    Sl   17:32   0:00 -rcd- -addr /tmp/nash.qNQa.sock
root         5  0.0  0.0  16028  3840 pts/4    S    17:32   0:00 /usr/bin/bash
root        23  0.0  0.0  34436  3056 pts/4    R+   17:34   0:00 ps aux
```

Everything except the `rfork` is like a dumb shell. Rfork will spawn a
new process with the namespace flags and executes the commands inside
the block on this namespace. It has the form:

```
rfork <flags> {
    <comands to run inside the container>
}
```

# OK, but what my deploy will look like?

Take a look in the script below:

```sh
#!/usr/bin/env nash

image="https://busybox.net/downloads/binaries/latest/busybox-x86_64"

-rm -rf rootfs

echo "Executing container"

# Forking the container with all namespaces except network
rfork upmis {
    mount -t proc proc /proc
    mkdir -p rootfs
    mount -t tmpfs -o size=1G tmpfs rootfs

    cd rootfs

    wget $image -O busybox
    chmod +x busybox

    mkdir bin

    ./busybox --install ./bin

    mkdir -p proc
    mkdir -p dev
    mount -t proc proc proc
    mount -t tmpfs tmpfs dev

    cp ../my-service .
    chroot . /my-service
}
```

Execute with:

```sh

./nash -file example.sh
--2016-04-15 17:54:02--  https://busybox.net/downloads/binaries/latest/busybox-x86_64
Resolving busybox.net (busybox.net)... 140.211.167.224
Connecting to busybox.net (busybox.net)|140.211.167.224|:443... connected.
HTTP request sent, awaiting response... 200 OK
Length: 973200 (950K)
Saving to: ‘busybox’

busybox                 100%[===============================>] 950.39K  21.1KB/s    in 43s

2016-04-15 17:54:46 (22.1 KB/s) - ‘busybox’ saved [973200/973200]

```

Change the last line of chroot to invoke /bin/sh if you want a shell
inside the busybox.

I know, I know, lots of questions in how to handle the hard parts of
deploy. My answer is: Coming soon.

# Didn't work?

I've tested in the following environments:

    Linux 4.1.13 (amd64)
    Fedora release 23

    Linux 4.3.3 (amd64)
    Archlinux

    Linux 4.1.13 (amd64)
    Debian 8

# Language specification

The specification isn't complete yet, but can be found
[here](https://github.com/tiago4orion/nash/blob/master/spec.ebnf).
The file `spec_test.go` makes sure it is sane.

# Want to contribute?

Open issues and PR :)
The project is in an early stage, be patient because things can change
a lot in the future.
