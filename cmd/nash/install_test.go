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
			libfiles: []string{
				"/testfile/file.sh",
			},
			installpath: "/testfile",
			want : map[string]string{
				"/testfile/file.sh" : "/testfile/file.sh",
			},
		},
		{
			name: "SingleDirWithMultipleFiles",
			libfiles: []string{
				"/testfile/file.sh",
				"/testfile/fileagain.sh",
			},
			installpath: "/testfile",
			want : map[string]string{
				"/testfile/file.sh" : "/testfile/file.sh",
				"/testfile/fileagain.sh" : "/testfile/fileagain.sh",
			},
		},
		{
			name: "MultipleDirsWithMultipleFiles",
			libfiles: []string{
				"/testfile/file.sh",
				"/testfile/dir1/file.sh",
				"/testfile/dir1/fileagain.sh",
				"/testfile/dir2/file.sh",
				"/testfile/dir2/fileagain.sh",
				"/testfile/dir2/dir3/file.sh",
			},
			installpath: "/testfile",
			want : map[string]string{
				"/testfile/file.sh": "/testfile/file.sh",
				"/testfile/dir1/file.sh" : "/testfile/dir1/file.sh",
				"/testfile/dir1/fileagain.sh": "/testfile/dir1/fileagain.sh",
				"/testfile/dir2/file.sh": "/testfile/dir2/file.sh",
				"/testfile/dir2/fileagain.sh": "/testfile/dir2/fileagain.sh",
				"/testfile/dir2/dir3/file.sh" : "/testfile/dir2/dir3/file.sh",
			},
		},
		{
			name: "InstallOnlyFilesIndicatedByInstallDir",
			libfiles: []string{
				"/testfile/file.sh",
				"/testfile/dir1/file.sh",
				"/testfile/dir1/fileagain.sh",
				"/testfile/dir2/file.sh",
				"/testfile/dir2/fileagain.sh",
				"/testfile/dir2/dir3/file.sh",
			},
			installpath: "/testfile/dir2",
			want : map[string]string{
				"/dir2/file.sh": "/testfile/dir2/file.sh",
				"/dir2/fileagain.sh": "/testfile/dir2/fileagain.sh",
				"/dir2/dir3/file.sh" : "/testfile/dir2/dir3/file.sh",
			},
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
			
			listNashPathFiles := func() []string {
				files := []string{}
				filepath.Walk(nashpath, func(path string, stats os.FileInfo, err error) error {
					if stats.IsDir() {
						return nil
					}
					files = append(files, path)
					return nil
				})
				return files
			}
			
			gotFiles := listNashPathFiles()
			
			fatal := func() {
				t.Errorf("nashpath: [%s]", nashpath)
				t.Errorf("nashpath contents:")
				
				for _, path := range gotFiles {
					t.Errorf("[%s]", path)
				}
				t.Fatal("")
			}
			
			if len(gotFiles) != len(c.want) {
				t.Errorf("wanted[%d] files but got[%d]", len(c.want), len(gotFiles))
				fatal()
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
					t.Errorf("error[%s] checking wanted file[%s]", err, wantFilepath)
					fatal()
				}
				gotContentsRaw, err := ioutil.ReadAll(wantFile)
				wantFile.Close()
				
				if err != nil {
					t.Errorf("error[%s] checking existence of wanted file[%s]", err, wantFilepath)
					fatal()
				}
				
				gotContents := string(gotContentsRaw)
				if gotContents != wantContents {
					t.Errorf("for file [%s] wanted contents [%s] but got [%s]", wantFilepath, wantContents, gotContents)
					fatal()
				}
			}
		})
	}
}

func TestSourcePathCantBeEqualToNashLibDir(t *testing.T) {
		nashpath, rmnashpath := fixture.Tmpdir(t)
		defer rmnashpath()
		
		nashlibdir := main.NashLibDir(nashpath)
		
		fixture.CreateFile(t, filepath.Join(nashlibdir, "whatever.sh"))
		
		assertInstallLibFails(t, nashpath, nashlibdir)
}

func TestSourcePathCantBeInsideNashLibDir(t *testing.T) {
		nashpath, rmnashpath := fixture.Tmpdir(t)
		defer rmnashpath()
		
		nashlibdir := main.NashLibDir(nashpath)
		sourcelibdir := filepath.Join(nashlibdir, "somedir")
		fixture.CreateFile(t, filepath.Join(sourcelibdir, "whatever.sh"))
		
		assertInstallLibFails(t, nashpath, sourcelibdir)
}

func TestRelativeSourcePathCantBeInsideNashLibDir(t *testing.T) {
		nashpath, rmnashpath := fixture.Tmpdir(t)
		defer rmnashpath()
		
		nashlibdir := main.NashLibDir(nashpath)
		fixture.CreateFile(t, filepath.Join(nashlibdir, "somedir", "whatever.sh"))
		
		oldwd := fixture.WorkingDir(t)
		defer fixture.ChangeDir(t, oldwd)
		
		fixture.ChangeDir(t, nashlibdir)
		
		assertInstallLibFails(t, nashpath, "./somedir")
}

func TestFailsOnUnexistentSourcePath(t *testing.T) {
		nashpath, rmnashpath := fixture.Tmpdir(t)
		defer rmnashpath()
		
		assertInstallLibFails(t, nashpath, "/nonexistent/nash/crap")
}

func TestFailsOnUnreadableSourcePath(t *testing.T) {
		nashpath, rmnashpath := fixture.Tmpdir(t)
		defer rmnashpath()
		
		sourcedir, rmsourcedir := fixture.Tmpdir(t)
		defer rmsourcedir()
		
		fixture.Chmod(t, sourcedir, 0)
		assertInstallLibFails(t, nashpath, sourcedir)
}

func TestFailsOnUnreadableFileInsideSourcePath(t *testing.T) {
		nashpath, rmnashpath := fixture.Tmpdir(t)
		defer rmnashpath()
		
		sourcedir, rmsourcedir := fixture.Tmpdir(t)
		defer rmsourcedir()
		
		readableFile := filepath.Join(sourcedir, "file1.sh")
		unreadableFile := filepath.Join(sourcedir, "file2.sh")
		
		fixture.CreateFiles(t, []string{readableFile, unreadableFile})
		fixture.Chmod(t, unreadableFile, 0) 
		
		assertInstallLibFails(t, nashpath, sourcedir)
}

func assertInstallLibFails(t *testing.T, nashpath string, sourcepath string) {
		t.Helper()
		
		err := main.InstallLib(nashpath, sourcepath)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
}