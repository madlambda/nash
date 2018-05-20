package main_test

import (
	"os"
	"io/ioutil"
	"strings"
	"testing"
	"path/filepath"
	
	"github.com/NeowayLabs/nash/cmd/nash"
	"github.com/NeowayLabs/nash/internal/testing/fixture"
)

// TODO: test when nashpath lib already exists and has libraries inside
// TODO: test when you install a lib using a path that is inside nashpath

func TestInstallLib(t *testing.T) {

	type testcase struct {
		name string
		libfiles []string
		libpath string
	}
	
	cases := []testcase{
		{
			name: "Dir",
		},
		{
			name: "File",
				libfiles: []string{
				"/testfile/file.sh",
			},
			libpath: "/testfile",
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
			name: "FilesAndDirsRecursively",
		},
	}
	
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			nashpath, rmnashpath := fixture.Tmpdir(t)
			defer rmnashpath()
			
			libfilesDir, rmlibfilesDir := fixture.Tmpdir(t)
			defer rmlibfilesDir()
			
			nashlibdir := main.NashLibDir(nashpath)
			libfiles := []string{}
			
			for _, f := range c.libfiles {
				libfiles = append(libfiles, filepath.Join(libfilesDir, f))
			}
			
			createdLibFiles := fixture.CreateFiles(t, libfiles)
			wantedFiles := map[string]string{}
			
			for createdFile, fileContents := range createdLibFiles {
				wantedFilepath := strings.TrimPrefix(createdFile, libfilesDir)
				if !strings.HasPrefix(wantedFilepath, c.libpath) {
					continue
				}
				wantedFilepath = filepath.Join(nashlibdir, wantedFilepath)
				wantedFiles[wantedFilepath] = fileContents
			}
			
			libpath := filepath.Join(libfilesDir, c.libpath)
			err := main.InstallLib(nashpath, libpath)
			if err != nil {
				t.Fatal(err)
			}
			
			for wantFilepath, wantContents := range wantedFiles {
				wantFile, err := os.Open(wantFilepath)
				if err != nil {
					t.Fatalf("error[%s] opening wanted file[%s]", err, wantFilepath)
				}
				gotContentsRaw, err := ioutil.ReadAll(wantFile)
				wantFile.Close()
				
				if err != nil {
					t.Fatalf("error[%s] checking existence of wanted file[%s]", err, wantFilepath)
				}
				
				gotContents := string(gotContentsRaw)
				if gotContents != wantContents {
					t.Fatalf("for file [%s] wanted contents [%s] but got [%s]", wantFilepath, wantContents, gotContents)
				}
			}
		})
	}
}