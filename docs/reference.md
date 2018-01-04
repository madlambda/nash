<!-- mdtocstart -->

# Table of Contents

- [Command line arguments](#command-line-arguments)
- [Flow control](#flow-control)
    - [Branching](#branching)
    - [Looping](#looping)
        - [Lists](#lists)
        - [Forever](#forever)
- [Functions](#functions)
- [Operators](#operators)
    - [+](#)
        - [string](#string)
- [Packages](#packages)
- [Iterating](#iterating)
- [Built-in functions](#builtin-functions)
    - [print](#print)
    - [format](#format)
    - [len](#len)
    - [append](#append)
    - [exit](#exit)
    - [glob](#glob)
- [Standard Library](#standard-library)

<!-- mdtocend -->

Here lies a comprehensive reference documentation of nash
features and built-in functions, and how to use them.

There is also some [examples](./examples) that can be useful.


# Command line arguments

To handle script arguments you can just use the ARGS variable,
that is a list populated with the arguments passed to your script
when it is executed, like:

```nash
echo
echo "acessing individual parameter"
var somearg = $ARGS[0]
echo $somearg
echo
```

# Flow control

## Branching

To branch you can use **if** statement, it requires
a boolean expression, like the comparison operator:

```nash
var a = "nash"
echo -n $a
if $a == "nash" {
    a = "rocks"
}
echo $a
#Output:"nashrocks"
```

You can also use a junction of boolean expressions:

```nash
a = "nash"
b = "rocks"
if $a == "nash" && $b == "rocks"{
    echo "hellyeah"
}
#Output:"hellyeah"
```

You can also use a disjunction of boolean expressions:

```nash
a = "nash"
b = "rocks"
if $a == "bash" || $b == "rocks"{
    echo "hellyeah"
}
#Output:"hellyeah"
```

## Looping

Right now there are two kind of loops, on lists
and the forever kind :-).

### Lists

You can iterate lists like this:

```nash
a = ""
for i in ("nash" "rocks"){
    a = $a + $i
}
echo $a
#Output:"nashrocks"
```

### Forever

It would be cool to loop on boolean expressions, but
right now we can only loop forever (besides list
looping):

```nash
for {
    echo "hi"
}
```

# Functions

Defining functions is very easy, for example:

```nash
fn concat(a, b) {
        return $a+$b
}

res <= concat("1","9")
echo $res

#Output:"19"
```

If a parameter is missing on the function call,
it will fail:

```nash
fn concat(a, b) {
        return $a, $b
}

res <= concat("1")
echo $res

#Output:"ERROR: Wrong number of arguments for function concat. Expected 2 but found 1"
```

Passing extra parameters will also fail:

```nash
fn concat(a, b) {
        return $a, $b
}

res <= concat("1","2","3")
echo $res

#Output:"ERROR: Wrong number of arguments for function concat. Expected 2 but found 3"
```

# Operators

## +

The **+** operator behaviour
is dependent on the type its operands. It
is always invalid to mix types on the operation
(like one operand is a string and the other one is a integer).

The language is dynamically typed, but it is strongly
typed, types can't be mixed on operations, there is no
implicit type coercion.

### string

String concatenation is pretty straightforward.
For example:

```nash
a = "1"
b = "2"

echo $a+$b
#Output:"12"
```

# Packages

TODO

# Iterating

TODO

# Built-in functions

Built-in functions are functions that are embedded on the
language. You do not have to import any package to use them.

## print

The function **print** is used to print simple
messages directly to stdout:

```nash
print("hi")
#Output:"hi"
```

And supports formatting:

```nash
print("%s:%s", "1", "2")
#Output:"1:2"
```

## format

The function **format** is used like **print**, but
instead of writing to stdout it will return the string
according to the format provided:

```nash
a <= format("%s:%s", "1", "2")
echo $a
#Output:"1:2"
```

## len

The function **len** returns the length of a list.
An example to check for the length of a list:

```
echo "define one list with two elements"
args = (
    "one"
    "two"
)
echo "getting list length"
argslen <= len($args)
echo $argslen
```

## append

The function **append** appends one element to the end of a exist list.
Append returns the updated list.

An example to append one element to a exist list:

```
example_list = ()
echo "appending string 1"
example_list <= append($example_list, "1")
echo $example_list
echo "appending string 2"
example_list <= append($example_list, "2")
echo $example_list
```

## exit

TODO

## glob

TODO

# Standard Library

The standard library is a set of packages that comes with the
nash install (although not obligatory).

They must be imported explicitly (as any other package) to
be used.

* [fmt](docs/stdlib/fmt.md)
