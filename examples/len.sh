#!/usr/bin/env nash

echo "args: "
echo $ARGS

if len($ARGS) == "1" {
        echo "one parameter passed"
} else {
        echo "more parameters passed"
}
