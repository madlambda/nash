#!/usr/bin/env nash

echo "iterating through the arguments list"
echo ""
for arg in $ARGS {
	echo $arg
}

echo ""
echo "acessing individual parameter"
somearg = $ARGS[0]
echo $somearg
echo ""
