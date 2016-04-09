# cnt

[![Build Status](https://travis-ci.org/tiago4orion/cnt.svg?branch=master)](https://travis-ci.org/tiago4orion/cnt) [![codecov.io](https://codecov.io/github/tiago4orion/cnt/coverage.svg?branch=master)](https://codecov.io/github/tiago4orion/cnt?branch=master)

Simple examples of Linux namespaces in Go. Basically a rewrite of C examples in the article "Namespaces in Operation" of lwn.net[1].

Because Go isn't low level enought for some namespace's syscalls, we
use some hacks to circunvent the problem.

1. https://lwn.net/Articles/531114/
