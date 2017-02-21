#!/usr/bin/env nash

word = $ARGS[1]
sep =$ARGS[2]

fn splitter(char) {
        if $char == $sep {
            return "0"
        }
        return "1"
}

output <= split($word, $splitter)

for o in $output {
	echo $o
}
