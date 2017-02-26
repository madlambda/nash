#!/usr/bin/env nash

word = $ARGS[1]
sep = $ARGS[2]
output <= split($word, $sep)
for o in $output {
	echo $o
}
