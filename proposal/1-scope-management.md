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
a new variable **l**. That is why the **get** and **print** functions
are always referencing an outer list **l** that is empty, a new one
is created each time the add function is called.

In this document we brainstorm about possible solutions to this.

## Proposal I - "var"

This proposal adds a new keyword `var` that will be used to declare and
initialize variables in the local scope. Is an error to use `var` with
an existent local variable (redeclare is forbiden).

```js
var i = "0"
```

Normal assignments will only update existent variables. The assignment
must first look for the target variable in the local scope and then in
the parent, recursively, until it's found and then updated, otherwise
(in case the variable is not found) the interpreter must abort with
error.

```sh
var count = "0" # declare local variable

fn inc() {
	# update outer variable
	count, _ <= expr $count "+" 1
}

inc()
print($count) 	# outputs: 2
```

Below is how this proposal solves the scope management problem example:

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

Sintactically, the `var` statement is an extension of the assignment
and exec-assignment statements, and then it should support multiple
declarations in a single statement also. Eg.:

```sh
var i, j = "0", "1"

var body, err <= curl -f $url

var name, surname, err <= getAuthor()
```

One of the downsides of `var` is the requirement that none of the
targeted variable exists, because it makes awkward when existent
variables must be used in conjunction with new ones. An example is the
variables `$status` and `$err` that are often used to get process exit
status and errors from functions, respectively.

The [PR #227](https://github.com/NeowayLabs/nash/pull/227) implements
this proposal but deviates in multiple assignments to handle the
downside above. The `var` statement was implemented with the rules
below:

1. At least one of the targeted variables must do not exists;
2. The existent variables are just updated in the scope it resides;

Below are some valid examples with [#227](https://github.com/NeowayLabs/nash/pull/227):

```sh
var a, b = "0", "1" # works fine, variables didn't existed before

var a, b = "2", "3" # error by rule 1

# works! c is declared but 'a' and 'b' are updated (by rule 2)
var a, b, c = "4", "5", "6"

# works, variables first declared
var users, err <= cat /etc/passwd | awk -F ":" "{print $1}"

# also works, but $err just updated
var pass, err <= cat /etc/shadow | awk -F ":" "{print $2}"
```

The implementation above is handy but makes the meaning of `var`
confuse because it declares new variables **and** update existent ones
(in outer scopes also). Then making hard to know what variables are
being declared local and what are being updated, by just looking at
the statement, because the meaning will depend in the current
environment of variables.

Another downside of `var` is their very incompatible nature. Every
nash script ever created will be affected.

## Proposal II - "outer"

This proposal adds a new `outer` keyword that permits the update of
variables in the outer scope. Outer assignments with non-existent
variables is an error.

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

	fn not_clear() {
		# "l" is not cleared, but a new a new variable is created (shadowing)
		# because "outer" isn't specified.
		l = ()
	}

	return $add, $get, $string
}
```

The `outer` keyword has the same meaning that Python's `global`
keyword.
