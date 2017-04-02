<!-- mdtocstart -->
# Table of Contents

- [Command line arguments](#command-line-arguments)
- [Flow control](#flow-control)
- [Functions](#functions)
- [Operators](#operators)
    - [+ (Concatenation)](#-concatenation)
        - [string](#string)
- [Iterating](#iterating)
- [Built-in functions](#builtin-functions)
    - [len](#len)
    - [append](#append)
- [Packages](#packages)
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
somearg = $ARGS[0]
echo $somearg
echo
```

# Flow control

## Branching

To branch you can use **if** statement, it requires
a boolean expression, like the comparison operator:

```nash
a = "nash"
echo $a
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

The **+** is the concatenation operator. Its behaviour
is dependent on the type it is concatenating. It
is always invalid to mix types on the operation
(like concatenating a string with a integer).

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

# Iterating

# Built-in functions

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

# Packages
