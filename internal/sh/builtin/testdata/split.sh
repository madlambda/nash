#!/usr/bin/env nash

var word = $ARGS[1]
var sep = $ARGS[2]
var output <= split($word, $sep)
for o in $output {
	echo $o
}
