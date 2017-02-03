# Reference

Here lies a comprehensive reference documentation of nash
features and built in features, and how to use them.


## Script args

To handle script arguments you can just use the ARGS variable,
that is a list populated with the arguments passed to your script
when it is executed, like:

```
echo ""
echo "acessing individual parameter"
somearg = $ARGS[0]
echo $somearg
echo ""
```

## Built-in functions

### len

The function **len** returns the length of a list.
An example to check for the length of a list:

```
echo "define one list with two elemnts"
args = (
    "one"
    "two"
)
echo "getting list length"
argslen <= len($args)
echo $argslen
```

### append

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




