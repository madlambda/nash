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
			name: "Dir",
		},
		{
			name: "File",
		},
		{
			name: "Dirs",
		},
		{
			name: "Files",
		},
		{
			name: "DirsRecursively",
		},
		{
			name: "FileAndDirRecursively",
		},
	}
	
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
		})
	}
}