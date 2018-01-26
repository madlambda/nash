package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	const defaultMinTextSize = 6
	var minTextSize uint

	flag.UintVar(
		&minTextSize,
		"s",
		defaultMinTextSize,
		"the minimum size in runes to characterize as a text",
	)

	scanner := Do(os.Stdin, minTextSize)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if scanner.Err() != nil {
		fmt.Printf("error: %s", scanner.Err())
		os.Exit(1)
	}
}
