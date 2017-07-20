#!/usr/bin/env nash

print("iterating through the arguments list\n\n")
for arg in $ARGS {
    print("%s\n", $arg)
}

print("\n")
print("acessing individual parameter\n")
var somearg = $ARGS[0]
print("%s\n", $somearg)
