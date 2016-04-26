package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintf(os.Stdout, "stdout...\n")
	fmt.Fprintf(os.Stderr, "stderr...\n")
}
