# nash

[![Build Status](https://travis-ci.org/NeowayLabs/nash.svg?branch=master)](https://travis-ci.org/NeowayLabs/nash) [![codecov.io](https://codecov.io/github/NeowayLabs/nash/coverage.svg?branch=master)](https://codecov.io/github/NeowayLabs/nash?branch=master)

Nash is a Linux system shell that attempts to be more safe and give
more power to user. It's safe in the sense of it's far more hard to
shoot yourself in the foot, (or in the head). It gives
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
configurations in every machine was maintained by several (sometimes
un-mantainable) scripts. Today we know that this approach have lots of
problems and the container approach is a better alternative. But in the
other end, we're paying a high cost for the lose of control. The
container-technologies in the market are very unsafe and few people are
worrying about. No one knows for hundred percent sure, how the things
really works because after every release it's done differently. On my
view it's getting worse and worse...

Before Linux namespace, BSD Jails, Solaris Zones, and so on, the
sysadmin had to fight the global view of the operating
system. There was only one root mount table, only one view of devices
and processes, and so on. It was a mess. This approach then proved to
be much harder to scale because of the services conflicts (port numbers,
files on disk, resource exhaustion, etc) in the global interface.
The container/namespace idea creates an abstraction to the process in
a way that it thinks it's the only process running (not counting init),
it is the root and then, the filesystem of the container only has the files
required for it (nothing more).

What's missing is a safe and robust shell for natural usage of namespace/container ideas
for everyone (programmers, sysadmins, etc).

Nasn is a way for you, that understand the game rules, to make
reliable deploy scripts using the good parts of the container
technologies. If you are a programmer, you can use a good language to
automate the devops instead of relying on lots of different
technologies (docker, rkt, k8s, mesos, terraform, and so on). And you
can create libraries for code-reuse.

It's only a simple shell plus a keyword called `rfork`. Rfork try
to mimic what Plan9 `rfork` does for namespaces, but with linux
limitations in mind.

# Show time!

Go ahead:

```sh
go get github.com/NeowayLabs/nash/cmd/nash
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
λ> id
uid=1000(user) gid=1000(user) groups=1000(user),98(qubes)
λ> rfork u {
     id
}
uid=0(root) gid=0(root) groups=0(root),65534
```
Yes, Linux supports creation of containers by unprivileged users. Tell
this to the customer success of your container-infrastructure-vendor. :-)

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
types of namespaces).  If you want another shell (maybe bash) inside
the namespace:

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

Everything except the `rfork` is like a common shell. Rfork will spawn a
new process with the namespace flags and executes the commands inside
the block on this namespace. It has the form:

```sh
rfork <flags> {
    <statements to run inside the container>
}
```

Nash stops executing the script at first error found. Commands have an explicitly
way to bypass such restriction by prepending a dash '-' to the command statement.
For example:

```sh
λ> rm file-not-exists
rm: cannot remove ‘file-not-exists’: No such file or directory
ERROR: exit status 1
λ> -rm file-not-exists
rm: cannot remove ‘file-not-exists’: No such file or directory
λ>
```
The dash '-' works only for OS commands, other kind of errors are impossible to bypass.

```sh
λ> echo $PATH
/bin:/sbin:/usr/bin:/usr/local/bin:/home/user/.local/bin:/home/user/bin:/home/user/.gvm/pkgsets/go1.5.3/global/bin:/home/user/projects/3rdparty/plan9port/bin:/home/user/.gvm/gos/go1.5.3/bin
λ> echo $bleh
ERROR: Variable '$bleh' not set
```
# OK, but how scripts should look like?

Take a look in the script below:

```sh
#!/usr/bin/env nash
#
# Execute `my-service` inside a busybox container

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
[here](https://github.com/NeowayLabs/nash/blob/master/spec.ebnf).
The file `spec_test.go` makes sure it is sane.

# Bash comparison

| Bash                            	| Nash                                  	| Description                                                                      	|
|---------------------------------	|---------------------------------------	|----------------------------------------------------------------------------------	|
| `GOPATH=/home/user/gopath`        	| `GOPATH="/home/user/gopath"`            	| Nash enforces quoted strings                                                     	|
| `GOPATH="$HOME/gopath"`           	| `GOPATH=$HOME+"/gopath"`                	| Nash doesn't do string expansion                                                 	|
| `export PATH=/bin:/usr/bin`       	| `PATH="/bin:/usr/bin"`<br>`setenv PATH`      	| setenv operates only on valid variables                                          	|
| export                          	| showenv                               	|                                                                                  	|
| ls -la                          	| ls -la                                	| Simple commads are identical                                                     	|
| ls -la "$GOPATH"                	| ls -la $GOPATH                        	| Nash variables shouldn't be enclosed in quotes, because it's default behaviour. 	|
| ./worker -d 2>log.err 1>log.out 	| ./worker -d >[2] log.err >[1] log.out 	| Nash redirection works like plan9 rc                                             	|
| ./worker -d 2>&1                	| ./worker -d >[2=1]                    	| Redirection map only works for standard file descriptors (0,1,2)                 	|


# Security

The PID 1 of every namespace created by `nash` is the same nash binary reading
commands from the parent shell via unix socket. It allows the parent namespace
(the script that creates the namespace) to issue commands inside the child
namespace. In the current implementation the unix socket communication is not
secure yet.

# Motivation

I needed to create test scripts to be running on different mount namespaces
for testing a file server and various use cases. Using bash in addition to
docker or rkt was not so good for various reasons. First, docker prior to version 1.10
doesn't support user namespaces, and then my `make test` would requires root privileges,
but for docker 1.10 user namespace works still requires to it being enabled in the
daemon flags (--userns-remap=?) making more hard to work on standard CIs (travis, circle, etc)...
Another problem was that it was hard to maintain a script, that spawn docker container
scripts inheriting environment variables from parent namespace (or host). Docker treats the container as a different
machine or VM, even calling the parent namespace as "host". This breaks the namespace
sharing/unsharing idea of processes. What I wanted was a copy of the missing plan9
environment namespace to child namespaces.

# Want to contribute?

Open issues and PR :)
The project is in an early stage, be patient because things can change in the future.
