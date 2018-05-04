package nash_test

import (
	"path/filepath"
	
	"github.com/NeowayLabs/nash"
)

func Example() {

	nashpath := filepath.Join("tmp", "nashpath")
	nashroot := filepath.Join("tmp", "nashroot")

	nash, err := nash.New(nashpath, nashroot)

	if err != nil {
		panic(err)
	}

	// Execute a script from string
	err = nash.ExecuteString("-input-", `echo Hello World`)

	if err != nil {
		panic(err)
	}

	// Output: Hello World
}
