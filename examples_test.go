package nash_test

import (
	"os"
	"io/ioutil"
	
	"github.com/madlambda/nash"
)

func Example() {

	nashpath,cleanup := tmpdir()
	defer cleanup()
	
	nashroot, cleanup := tmpdir()
	defer cleanup()

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

func tmpdir() (string, func()) {	
	dir, err := ioutil.TempDir("", "nash-tests")
	if err != nil {
		panic(err)
	}
	
	return dir, func() {
		err := os.RemoveAll(dir)
		if err != nil {
			panic(err)
		}
	}
}

