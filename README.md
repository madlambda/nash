# nash

[![Join the chat at https://gitter.im/NeowayLabs/nash](https://badges.gitter.im/NeowayLabs/nash.svg)](https://gitter.im/NeowayLabs/nash?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [![GoDoc](https://godoc.org/github.com/NeowayLabs/nash?status.svg)](https://godoc.org/github.com/NeowayLabs/nash) 
[![Build Status](https://travis-ci.org/NeowayLabs/nash.svg?branch=master)](https://travis-ci.org/NeowayLabs/nash) [![Go Report Card](https://goreportcard.com/badge/github.com/NeowayLabs/nash)](https://goreportcard.com/report/github.com/NeowayLabs/nash)

Nash is a system shell, inspired by plan9 `rc`, that makes it easy to create reliable and safe scripts taking advantages of operating systems namespaces (on linux and plan9) in an idiomatic way.

# Show time!

[![asciicast](https://asciinema.org/a/6a1lqcllwctoaej6gzsxcuqo6.png)](https://asciinema.org/a/6a1lqcllwctoaej6gzsxcuqo6?speed=2&autoplay=true)

## Useful stuff

- nashfmt: Formats nash code (like gofmt) but no code styling defined yet.
- [nashcomplete](https://github.com/NeowayLabs/nashcomplete): Autocomplete done in nash script.
- [Dotnash](https://github.com/lborguetti/dotnash): Nash profile customizations (e.g: prompt, aliases, etc)
- [nash-mode](https://github.com/tiago4orion/nash-mode.el): Emacs major mode integrated with `nashfmt`.

## Why nash scripts are reliable?

1. Nash aborts at first non-success status of commands;
2. Nash aborts at first unbound variable;
3. It's possible to check the result status of every component of a pipe;
4. **no eval**;
5. Strings are pure strings (no evaluation of variables);
6. No wildcards (globbing) of files; ('rm \*' removes a file called '\*');
7. No [obscure](http://explainshell.com/) syntax;
8. Support tooling for indent/format and statically analyze the scripts;

## Installation

If you have Go, go-get it:

```sh
# Make sure GOPATH/bin is in your PATH
go get github.com/NeowayLabs/nash/cmd/nash
```
If not, [download the latest binary release](https://github.com/NeowayLabs/nash/releases) and copy to somewhere in your PATH.

## Getting started

Nash syntax resembles a common shell:

```
nash
λ> echo "hello world"
hello world
```
Pipes works like borne shell and derivations:

```sh
λ> cat spec.ebnf | wc -l
108
```
Output redirection works like Plan9 rc, but not only for filenames. It
supports output redirection to tcp, udp and unix network protocols.

```sh
# stdout to log.out, stderr to log.err
λ> ./daemon >[1] log.out >[2] log.err
# stderr pointing to stdout
λ> ./daemon-logall >[2=1]
# stdout to /dev/null
λ> ./daemon-quiet >[1=]
# stdout and stderr to tcp address
λ> ./daemon >[1] "udp://syslog:6666" >[2=1]
# stdout to unix file
λ> ./daemon >[1] "unix:///tmp/syslog.sock"
```

**For safety, there's no `eval` or `string/tilde expansion` or `command substitution` in Nash.**

To assign command output to a variable exists the '<=' operator. See the example
below:
```sh
fullpath <= realpath $path | xargs -n echo
echo $fullpath
```
The symbol '<=' redirects the stdout of the command or function invocation in the
right-hand side to the variable name specified.

If you want the command output splited into an array, then you'll need
to store it in a temporary variable and then use the builtin `split` function.

```sh
out <= find .
files <= split($out, "\n")

for f in $files {
        echo "File: " + $f
}
```

To avoid problems with spaces in variables being passed as multiple arguments to commands,
nash pass the contents of each variable as a single argument to the command. It works like
enclosing every variable with quotes before executing the command. Then the following example
do the right thing:
```sh
fullname = "John Nash"
./ci-register --name $fullname --option somevalue
```
On bash you need to enclose the `$fullname` variable in quotes to avoid problems.

Nash syntax does not support shell expansion from strings. There's no way to
do things like the following in nash:
```bash
echo "The date is: $(date +%D)" # DOESNT WORKS!
```
Instead you need to assign each command output to a proper variable and then concat
it with another string when needed. In nash, the example above must be something
like that:
```sh
today <= date "+%D"
echo "The date is: " + $today
```
The concat operator (+) could be used between variables and literal strings.

Functions can be declared with "fn" keyword:
```sh
fn cd(path) {
    fullpath <= realpath $path | xargs echo -n
    chdir($path)
    PROMPT="[" + $fullpath + "]> "
    setenv PROMPT
}
```

And can be invoked as a normal function invocation:
```sh
λ> cd("/etc")
[/etc]>
```
Functions are commonly used for nash libraries, but when needed it can be bind'ed
to some command name. Using the `cd` function below, we can override the builtin `cd`
with that command with `bindfn` statement.
```sh
λ> # bindfn syntax is:
λ> # bindfn <function-name> <cmd-name>
λ> bindfn cd cd
λ> cd /var/log
[/var/log]>
```
The only control statements available are `if`, `else` and `for`.
In the same way, nash doesn't support shell expansion at `if` condition.
For check if a directory exists you must use:
```sh
-test -d $rootfsDir    # if you forget '-', the script will be aborted here
                       # if path not exists

if $status != "0" {
        echo "RootFS does not exists."
        exit $status
}
```
Nash stops executing the script at first error found and, in the majority of times, it is what
you want (specially for deploys). But Commands have an explicitly
way to bypass such restriction by prepending a dash '-' to the command statement.
For example:

```sh
fn cleanup()
        -rm -rf $buildDir
        -rm -rf $tmpDir
}
```

The dash '-' works only for operating system commands, other kind of errors are impossible to bypass.
For example, trying to evaluate an unbound variable aborts the program with error.

```sh
λ> echo $PATH
/bin:/sbin:/usr/bin:/usr/local/bin:/home/user/.local/bin:/home/user/bin:/home/user/.gvm/pkgsets/go1.5.3/global/bin:/home/user/projects/3rdparty/plan9port/bin:/home/user/.gvm/gos/go1.5.3/bin
λ> echo $bleh
ERROR: Variable '$bleh' not set
```

Long commands can be splited in multiple lines:

```sh
λ> (aws ec2 attach-internet-gateway	--internet-gateway-id $igwid
									--vpc-id $vpcid)

λ> instanceId <= (
	aws ec2 run-instances
			--image-id ami-xxxxxxxx
			--count 1
			--instance-type t1.micro
			--key-name MyKeyPair
			--security-groups my-sg
    | jq ".Instances[0].InstanceId"
)
λ> echo $instanceId
```

# Accessing command line args

When you run a nash script like:

```
λ> nash ./examples/args.sh --arg value
```

You can get the args using the **ARGS** variable, that is a list:

```
#!/usr/bin/env nash

echo "iterating through the arguments list"
echo ""
for arg in $ARGS {
	echo $arg
}
```


# Namespace features

Nash is built with namespace support only on Linux (Plan9 soon). If
you use OSX, BSD or Windows, then the `rfork` keyword will fail.

*The examples below assume you are on a Linux box.*

Below are some facilities for namespace management inside nash.
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

# OK, but how scripts should look like?

See the project [nash-app-example](https://github.com/NeowayLabs/nash-app-example).

# Didn't work?

I've tested in the following environments:

    Linux 4.7-rc7
    Archlinux

    Linux 4.5.5 (amd64)
    Archlinux

    Linux 4.3.3 (amd64)
    Archlinux

    Linux 4.1.13 (amd64)
    Fedora release 23

    Linux 4.1.13 (amd64)
    Debian 8

# Language specification

The specification isn't complete yet, but can be found
[here](https://github.com/NeowayLabs/nash/blob/master/spec.ebnf).
The file `spec_test.go` makes sure it is sane.

# Some Bash comparisons

| Bash | Nash | Description	|
| --- | --- | --- |
| `GOPATH=/home/user/gopath` | `GOPATH="/home/user/gopath"` | Nash enforces quoted strings |
| `GOPATH="$HOME/gopath"` | `GOPATH=$HOME+"/gopath"` | Nash doesn't do string expansion |
| `export PATH=/usr/bin` | `PATH="/usr/bin"`<br>`setenv PATH` | setenv operates only on valid variables |
| `export` | `showenv` | |
| `ls -la` | `ls -la` | Simple commads are identical |
| `ls -la "$GOPATH"` | `ls -la $GOPATH` | Nash variables shouldn't be enclosed in quotes, because it's default behaviour |
| `./worker 2>log.err 1>log.out` | `./worker >[2] log.err >[1] log.out` | Nash redirection works like plan9 rc |
| `./worker 2>&1` | `./worker >[2=1]` | Redirection map only works for standard file descriptors (0,1,2) |

# Security

The PID 1 of every namespace created by `nash` is the same nash binary reading
commands from the parent shell via unix socket. It allows the parent namespace
(the script that creates the namespace) to issue commands inside the child
namespace. In the current implementation the unix socket communication is not
secure yet.

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
files on disk, resource exhaustion, etc) in the global OS interface.
The container/namespace idea creates an abstraction to the process in
a way that it thinks it's the only process running (not counting init),
it is the root (or no) and then, the filesystem of the container only
has the files required for it (nothing more).

What's missing is a safe and robust shell for natural usage of
namespace/container ideas for everyone (programmers, sysadmins, etc).

Nash is a way for you, that understand the game rules, to make
reliable deploy scripts using the good parts of the container
technologies. If you are a programmer, you can use a good language to
automate the devops instead of relying on lots of different
technologies (docker, rkt, k8s, mesos, terraform, and so on). And you
can create libraries for code-reuse.

It's only a simple shell plus a keyword called `rfork`. Rfork try
to mimic what Plan9 `rfork` does for namespaces, but with linux
limitations in mind.


# Motivation

I needed to create test scripts to be running on different mount
namespaces for testing a file server and various use cases. Using bash
in addition to docker or rkt was not so good for various
reasons. First, docker prior to version 1.10 doesn't support user
namespaces, and then my `make test` would requires root privileges,
but for docker 1.10 user namespace works still requires to it being
enabled in the daemon flags (--userns-remap=?) making more hard to
work on standard CIs (travis, circle, etc)...  Another problem was
that it was hard to maintain a script, that spawn docker containers
inheriting environment variables from parent namespace (or
host). Docker treats the container as a different machine or VM, even
calling the parent namespace as "host". This breaks the namespace
sharing/unsharing idea of processes. What I wanted was a copy of the
missing plan9 'environment namespace' to child namespaces.

# Want to contribute?

Open issues and PR :)
The project is in an early stage, be patient because things can change in the future.

> "What I cannot create, I do not understand."
>
> -- Richard Feynman
