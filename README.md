# cnt

Simple examples of Linux namespaces in Go. Basically a rewrite of C examples in the article "Namespaces in Operation" of lwn.net[1].

Because Go isn't low level enought for some namespace's syscalls, we
use some hacks to circunvent the problem.

1. https://lwn.net/Articles/531114/