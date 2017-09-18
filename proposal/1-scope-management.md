# Proposal: Proper scope management

## Abstract

Currently on nash there is no way to properly work
with closures because scope management is very limited.

Lets elaborate on the problem by implementing a
list object by instantiating a set of functions
that manipulates the same data.

```sh
fn list() {
	l = ()

	fn add(val) {
		l <= append($l, $val)
	}

	fn get(i) {
		return $l[$i]
	}

	fn string() {
		print("list: [%s]\n", $l)
	}

	return $add, $get, $string
}
```

The idea is to hide all list data behind these 3 functions
that will manipulate the same data. The problem is that today
this is not possible, using this code:

```sh
add, get, string <= list()

$add("1")
$add("2")
$string()

v <= $get("0")
echo $v
```

Will result in:

```
list: []
/tmp/test.sh:27:5: /tmp/test.sh:11:23: Index out of bounds. len($l) == 0, but given 0
```

As you can see, even when we call the **add** function the list
remains empty, why is that ? The problem is on the add function:

```sh
fn add(val) {
	l <= append($l, $val)
}
```

When we reference the **l** variable it uses the reference on the
outer scope (the empty list), but there is no way to express syntactically
that we want to change the list on the outer scope instead of creating
a new variable **l** (shadowing the outer **l**).

That is why the **get** and **print** functions
are always referencing an outer list **l** that is empty, a new one
is created each time the add function is called.

In this document we navigate the solution space for this problem.

## Proposal I - Create new variables explicitly

On this proposal new variable creation requires an explicit
syntax construction.

We could add a new keyword `var` that will be used to declare and
initialize variables in the local scope, like this:

```js
var i = "0"
```

While the current syntax:

```js
i = "0"
```

Will be assigning a new value to an already existent variable **i**.
The assignment will first look for the target variable in the local
scope and then in the parent, traversing the entire stack, until it's
found and then updated, otherwise (in case the variable is not found)
the interpreter must abort with error.

```sh
var count = "0" # declare local variable

fn inc() {
	# update outer variable
	count, _ <= expr $count "+" 1
}

inc()
print($count) 	# outputs: 1
```

Below is how this proposal solves the list example:

```sh
fn list() {
	# initialize an "l" variable in this scope
	var l = ()

	fn add(val) {
		# use the "l" variable from parent scope
		# find first in the this scope if not found
		# then find variable in the parent scope
		l <= append($l, $val)
	}

	fn get(i) {
		# use the "l" variable from parent scope
		return $l[$i]
	}

	fn string() {
		# use the "l" variable from parent scope
		print("list: [%s]\n", $l)
	}

	fn not_clear() {
		# force initialize a new "l" variable in this scope
		# because this the "l" list in the parent scope is not cleared
		var l = ()
	}

	return $add, $get, $string
}
```

Syntactically, the `var` statement is an extension of the assignment
and exec-assignment statements, and then it should support multiple
declarations in a single statement also. Eg.:

```sh
var i, j = "0", "1"

var body, err <= curl -f $url

var name, surname, err <= getAuthor()
```

Using var always creates new variables, shadowing previous ones,
for example:


```sh
var a, b = "0", "1" # works fine, variables didn't existed before

var a, b, c = "4", "5", "6" # works! too, creating new a, b, c
```

On a dynamic typed language there is very little difference between
creating a new var or just reassigning it since variables are just
references that store no type information at all. For example,
what is the difference between this:

```
var a = "1"
a = ()
```

And this ?

```
var a = "1"
var a = ()
```

The behavior will be exactly the same, there is no semantic error
on reassigning the same variable to a value with a different type,
so reassigning on redeclaring has no difference at all (although it
makes sense for statically typed languages).

Statements are evaluated in order, so this:

```
a = ()
var a = "1"
```

Is **NOT** the same as this:

```
var a = "1"
var a = ()
```

This is easier to understand when using closures, let's go
back to our list implementation, we had something like this:

```
var l = ()

fn add(val) {
        # use the "l" variable from parent scope
        # find first in the this scope if not found
        # then find variable in the parent scope
        l <= append($l, $val)
}
```

If we write this:

```
var l = ()

fn add(val) {
        # creates new var
        var l = ()
        # manipulates new l var
        l <= append($l, $val)
}
```

The **add** function will not manipulate the **l** variable from the
outer scope, and our list implementation will not work properly.

But writing this:

```
var l = ()

fn add(val) {
        # manipulates outer l var
        l <= append($l, $val)
        # creates new var that is useless
        var l = ()
}
```

Will work, since we assigned a new value to the outer **l**
before creating a new **l** var.

The approach described here is very similar to how variables
are handled in [Lua](https://www.lua.org/), with the exception
that Lua uses the **local** keyword, instead of var.

Also, Lua allows global variables to be created by default, on
Nash we prefer to avoid global stuff and produce an error when
assigning new values to variables that do not exist.

Summarizing, on this proposal creating new variables is explicit
and referencing existent variables on outer scopes is implicit.


## Proposal II - Manipulate outer scope explicitly

This proposal adds a new `outer` keyword that permits the update of
variables in the outer scope. The default and implicit behavior of
variable assignments is to always create a new variable.

Considering our list example:

```sh
fn list() {
	# initialize an "l" variable in this scope
	l = ()

	fn add(val) {
		# use the "l" variable from the parent
		outer l <= append($l, $val)
	}

	fn get(i) {
		# use the "l" variable from the parent outer l
		return $l[$i]
	}

	fn string() {
		# use the "l" variable from the parent outer l
		print("list: [%s]\n", $l)
	}

	return $add, $get, $string
}
```

The `outer` keyword has the same meaning that Python's `global`
keyword.

Different from Python global, outer must appear on all assignments,
like this:

```sh
fn list() {
	# initialize an "l" variable in this scope
	l = ()

	fn doubleadd(val) {
		outer l <= append($l, $val)
		outer l <= append($l, $val)
	}

	return $doubleadd
}
```

This would be buggy and only add once:

```sh
fn list() {
	# initialize an "l" variable in this scope
	l = ()

	fn doubleadd(val) {
		outer l <= append($l, $val)
		l <= append($l, $val)
	}

	return $doubleadd
}
```

Trying to elaborate more on possible combinations
when using the **outer** keyword we get at some hard
questions, like what does outer means on this case:

```
fn list() {
    # initialize an "l" variable in this scope
    l = ()
    fn doubleadd(val) {
        l <= append($l, $val)
        outer l <= append($l, $val)
    }
    return $doubleadd
}
```

Will outer just handle the reference on its own scope or
will it jump its own scope and manipulate the outer variable ?

The name outer implies that it will manipulate the outer scope,
bypassing its own current scope, but how do you read the outer
variable ? We would need to support something like:

```
fn list() {
    # initialize an "l" variable in this scope
    l = ()
    fn add(val) {
        l <= "whatever"
        outer l <= append(outer $l, $val)
    }
    return $doubleadd
}
```

It is like with outer we are bypassing the lexical semantics
of the code, the order of declarations is not relevant anymore
since you have a form of "goto" to jump the current scope.

## Comparing both approaches

As everything in life, the design space for how to handle
scope management is full of tradeoffs.

Making outer scope management explicit makes declaring
new variables easier, since you have to type less to
create new vars.

But managing scope using closures gets more cumbersome,
consider this nested closures with the **outer** keyword:

```sh
fn list() {
	l = ()

	fn add(val) {
		# use the "l" variable from the parent
		outer l <= append($l, $val)
		fn addagain() {
		        outer l <= append($l, $val)
		}
		return $addagain
	}

	return $add
}
```

And this one with **var** :

```sh
fn list() {
	var l = ()

	fn add(val) {
		# use the "l" variable from the parent
		l <= append($l, $val)
		fn addagain() {
		        l <= append($l, $val)
		}
		return $addagain
	}

	return $add
}
```

The **var** option requires more writing for the common
case of declaring new variables, but makes closures pretty
natural to write, you just manipulate the variables
that exists lexically on your scope, like you would do
inside a **if** or **for** block.

Thinking about cognition, it seems easier to write buggy code
by forgetting to add an **outer** on the code than forgetting
to add a **var** and by mistake manipulate an variable outside
the scope.

The decision to break if the variable does not exist also enhances
the **var** option as less buggy since no new variable will be
created if you forget the **var**, but lexically reachable variables
will be manipulated (this is ameliorated by the fact that we don't have
global variables).

If we go for **outer** it seems that we are going to write less,
but some code, involving closures, will be harder to read (and write).
Since code is usually read more than it is written it seems like a sensible
choice to optimize for readability and understandability than just
save a few keystrokes.

But any statements made about cognition are really hard to be
considered as a global truth, since all human beings are biased which makes
identification of common patterns of cognition really hard. But if software
design has any kind of goal, must be this =).
