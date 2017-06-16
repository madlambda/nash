#!/usr/bin/env nash

var word = $ARGS[1]
var sep = $ARGS[2]

fn splitter(char) {
	if $char == $sep {
		return "0"
	}

	return "1"
}

var output <= split($word, $splitter)

for o in $output {
	echo $o
}
