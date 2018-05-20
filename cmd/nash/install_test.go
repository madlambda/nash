package main_test

import (
	"os"
	"io/ioutil"
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
		installpath string
		// want will map the wanted files to the original files copied from the lib
		// the wanted files paths are relative to inside the nashpath lib dir.
		// the files need to be mapped to the original files because of content validation
		// when multiple files are installed.
		want map[string]string
	}
	
	cases := []testcase{
		{
			name: "SingleFile",
			libfiles: []string{
				"/testfile/file.sh",
			},
			installpath: "/testfile/file.sh",
			want : map[string]string{
				"file.sh" : "/testfile/file.sh",
			},
		},
		{
			name: "SingleDir",
		},
		{
			name: "Dirs",
		},
		{
			name: "DirsRecursively",
		},
		{
			name: "WontCreateEntireLibTree",
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
			
			libfileFullPath := func(libfilepath string) string {
				return filepath.Join(libfilesDir, libfilepath)
			}
			for _, f := range c.libfiles {
				libfiles = append(libfiles, libfileFullPath(f))
			}
			
			createdLibFiles := fixture.CreateFiles(t, libfiles)
	
			installpath := filepath.Join(libfilesDir, c.installpath)
			err := main.InstallLib(nashpath, installpath)
			if err != nil {
				t.Fatal(err)
			}
			
			for wantFilepath, libfilepath := range c.want {
			
				completeLibFilepath := libfileFullPath(libfilepath)
				wantContents, ok := createdLibFiles[completeLibFilepath]
				
				if !ok {
					t.Errorf("unable to find libfilepath[%s] contents on created lib files map[%+v]", completeLibFilepath, createdLibFiles)
					t.Fatal("this probably means a wrongly specified test case with wanted files that are not present on the libfiles")
				}
			
				fullWantFilepath := filepath.Join(nashlibdir, wantFilepath)
				wantFile, err := os.Open(fullWantFilepath)
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