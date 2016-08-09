package scanner_test

import (
	"fmt"

	"github.com/NeowayLabs/nash/scanner"
)

func Example() {
	lex := scanner.Lex("-input-", `echo "hello world"`)

	for tok := range lex.Tokens {
		fmt.Println(tok)
	}

	// Output:
	// (COMMAND) - pos: 0, val: "echo"
	// (STRING) - pos: 6, val: "hello worl"...
	// EOF

}
