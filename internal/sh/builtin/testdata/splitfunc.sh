#!/usr/bin/env nash

word = $ARGS[1]
fn splitter(w) {
        echo $w
}

output <= split($word, $splitter)

for o in $output {
	echo $o
}
