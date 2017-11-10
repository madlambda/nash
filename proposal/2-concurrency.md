# Proposal: Concurrency on Nash

There has been some discussion on how to provide concurrency to nash.
There is a [discussion here](https://github.com/NeowayLabs/nash/issues/224) 
on how concurrency could be added as a set of built-in functions.

As we progressed discussing it seemed desirable to have a concurrency
that enforced no sharing between concurrent functions. It eliminates
races and forces all communication to happen explicitly, and the
performance overhead would not be a problem to a high level language
as nash.

Converging to a no shared state between concurrent functions initiated
the idea of using the current rfork built-in as a means to express
concurrency on Nash. This would already be possible today, the idea
is just to make it even easier, specially the communication between
different concurrent processes.

This idea enables an even greater amount of isolation between concurrent
processes since rfork enables different namespaces isolation (besides memory),
but it has the obvious fallback of not being very lightweight.

Since the idea of nash is to write simple scripts this does not seem
to be a problem. If it is on the future we can create lightweight concurrent
processes (green threads) that works orthogonally with rfork.

The prototype for the new rfork would be something like this:

```sh
chan <= rfork [ns_param1, ns_param2] (chan) {
        //some code
}
```

The code on the rfork block does not have access to the
lexical outer scope but it receives as a parameter a channel
instance.

This channel instance can be used by the forked processes and
by the creator of the process to communicate. We could use built-in functions:

```sh
chan <= rfork [ns_param1, ns_param2] (chan) {
        cwrite($chan, "hi")
}

a <= cread($chan)
```

Or some syntactic extension:

```sh
chan <= rfork [ns_param1, ns_param2] (chan) {
        $chan <- "hi"
}

a <= <-$chan
```

Since this channel is meant only to be used to communicate with
the created process, it will be closed when the process exit:

```sh
chan <= rfork [ns_param1, ns_param2] (chan) {
}

# returns empty string when channel is closed
<-$chan
```

Fan out and fan in should be pretty trivial:

```sh
chan1 <= rfork [ns_param1, ns_param2] (chan) {
}

chan2 <= rfork [ns_param1, ns_param2] (chan) {
}

# waiting for both to finish
<-$chan1
<-$chan2
```
