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
