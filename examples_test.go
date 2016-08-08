package nash_test

import "github.com/NeowayLabs/nash"

func Example() {
	nash, err := nash.New()

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
