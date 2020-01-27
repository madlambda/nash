package scanner_test

import (
	"fmt"

	"github.com/madlambda/nash/scanner"
)

func Example() {
	lex := scanner.Lex("-input-", `echo "hello world"`)

	for tok := range lex.Tokens {
		fmt.Println(tok)
	}

	// Output:
	// IDENT
	// STRING
	// ;
	// EOF

}
