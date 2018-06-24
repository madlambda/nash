<!-- mdtocstart -->

# Table of Contents

- [nash](#nash)
- [Show time!](#show-time)
    - [Useful stuff](#useful-stuff)
    - [Why nash scripts are reliable?](#why-nash-scripts-are-reliable)
    - [Installation](#installation)
        - [Release](#release)
            - [Linux](#linux)
        - [Master](#master)
    - [Getting started](#getting-started)
- [Accessing command line args](#accessing-command-line-args)
- [Namespace features](#namespace-features)
- [OK, but how scripts should look like?](#ok-but-how-scripts-should-look-like)
- [Didn't work?](#didnt-work)
- [Language specification](#language-specification)
- [Some Bash comparisons](#some-bash-comparisons)
- [Security](#security)
- [Installing libraries](#installing-libraries)
- [Releasing](#releasing)
- [Want to contribute?](#want-to-contribute)

<!-- mdtocend -->

# nash

[![Join the chat at https://gitter.im/NeowayLabs/nash](https://badges.gitter.im/NeowayLabs/nash.svg)](https://gitter.im/NeowayLabs/nash?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [![GoDoc](https://godoc.org/github.com/NeowayLabs/nash?status.svg)](https://godoc.org/github.com/NeowayLabs/nash)
[![Build Status](https://travis-ci.org/NeowayLabs/nash.svg?branch=master)](https://travis-ci.org/NeowayLabs/nash) [![Go Report Card](https://goreportcard.com/badge/github.com/NeowayLabs/nash)](https://goreportcard.com/report/github.com/NeowayLabs/nash)

Nash is a system shell, inspired by plan9 `rc`, that makes it easy to create reliable and safe scripts taking advantages of operating systems namespaces (on linux and plan9) in an idiomatic way.

# Show time!

[![asciicast](https://asciinema.org/a/6a1lqcllwctoaej6gzsxcuqo6.png)](https://asciinema.org/a/6a1lqcllwctoaej6gzsxcuqo6?speed=2&autoplay=true)

## Useful stuff

- nashfmt: Formats nash code (like gofmt) but no code styling defined yet (see Installation section).
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
    - On windows, the terminal does the globbing when in interactive mode.
    - On unix there's libs/completions to achieve something similar.
7. No [obscure](http://explainshell.com/) syntax;
8. Support tooling for indent/format and statically analyze the scripts;

## Installation

Nash uses two environment variables: **NASHROOT** to find the standard nash library and **NASHPATH** to find libraries in general (like user's code).

It is important to have two different paths since this will allow you
to upgrade nash (overwrite nash stdlib) without risking lost your code.

If **NASHPATH** is not set, a default of $HOME/nash will be assumed 
($HOMEPATH/nash on windows).
If **NASHROOT** is not set, a default of $HOME/nashroot will be assumed 
($HOMEPATH/nashroot on windows).

The libraries lookup dir will be $NASHPATH/lib.
The standard library lookup dir will be $NASHROOT/stdlib.

After installing the nash binary will be located at $NASHROOT/bin.

### Release

Installing is so stupid that we provide small scripts to do it.
If your platform is not supported take a look at the existent ones
and send a MR with the script for your platform.

#### Linux

Run:

```
./hack/install/linux.sh
```

### Master

Run:

```
make install
```

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
supports output redirection to tcp, udp and unix network protocols 
(unix sockets are not supported on windows).

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
var fullpath <= realpath $path | xargs -n echo
echo $fullpath
```
The symbol '<=' redirects the stdout of the command or function invocation in the
right-hand side to the variable name specified.

If you want the command output splited into an array, then you'll need
to store it in a temporary variable and then use the builtin `split` function.

```sh
var out <= find .
var files <= split($out, "\n")

for f in $files {
        echo "File: " + $f
}
```

To avoid problems with spaces in variables being passed as
multiple arguments to commands, nash pass the contents of each
variable as a single argument to the command. It works like
enclosing every variable with quotes before executing the command.
Then the following example do the right thing:

```sh
var fullname = "John Nash"
./ci-register --name $fullname --option somevalue
```
On bash you need to enclose the `$fullname` variable in quotes to avoid problems.

Nash syntax does not support shell expansion from strings. There's no way to
do things like the following in nash:

```bash
echo "The date is: $(date +%D)" # DOESNT WORKS!
```

Instead you need to assign each command output to a proper variable and then
concat it with another string when needed (see the [reference docs](./doc/reference.md)).

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

Long commands can be split in multiple lines:

```sh
λ> (aws ec2 attach-internet-gateway	--internet-gateway-id $igwid
									--vpc-id $vpcid)

λ> var instanceId <= (
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

# Installing libraries

Lets say you have a nash library and you want to install it. For example you have
the following:

```
awesome/code.sh
```

And you want to install it so you can write code like this:

```
import awesome/code

code_do_awesome_stuff()
```

All you have to do is run:

```
nash -install ./awesome
```

Or:

```
nash -install /absolute/path/awesome
```

The entire awesome dir (and its subdirs) will be copied where nash
searches for libraries (dependent on environment variables).

This is the recommended way of installing nash libraries (althought
you can do it manually if you want).

Single files can also be installed as packages, for example:

```
nash -install ./awesome/code.sh
```

Will enable you to import like this:

```
import code
```

If there is already a package with the given name it will be
overwritten.


# Releasing

To generate a release basically:

* Generate the release on github
* Clone the generated tag
* Run: ``` make release "version=<version>" ```

Where **<version>** must match the version of the git tag.

# Want to contribute?

Open issues and PR :)
The project is in an early stage, be patient because things can change in the future.

> "What I cannot create, I do not understand."
>
> -- Richard Feynman
