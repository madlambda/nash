#!/usr/bin/env nash

echo $ARGS
argslen <= len($ARGS)

test $argslen + "= 1"

if $status == "0" {
        echo "one parameter passed"
} else {
        echo "more parameters passed"
}
