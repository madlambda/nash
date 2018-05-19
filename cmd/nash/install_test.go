package main_test

import (
	"testing"
)

func TestInstallLib(t *testing.T) {

	type testcase struct {
		name string
	}
	
	cases := []testcase{
		{
			name: "InstallLibDir",
		},
		{
			name: "InstallLibFile",
		},
		{
			name: "InstallLibDirRecursively",
		},
	}
	
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
		})
	}
}