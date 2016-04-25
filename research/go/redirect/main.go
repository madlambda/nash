// Redirect is an experiment for execute:
// ls > file.out
// ls >[1] file.txt
// ls >[2] file.err >[1] file.out
// ls >[2=]
// ls >[2=1] file.out
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type (
	FDMap   map[*os.File]*os.File
	mapDecl struct {
		leftFD  int
		rightFD int
		out     string
	}
)

const (
	FDUnused  = -1
	FDSupress = -2
)

var (
	defFDmap map[int]*os.File = map[int]*os.File{
		0: os.Stdin,
		1: os.Stdout,
		2: os.Stderr,
	}
)

func buildRedirect(redirDecl mapDecl, m FDMap) (FDMap, error) {
	if defFDmap[redirDecl.leftFD] != nil {
		if defFDmap[redirDecl.rightFD] != nil {
			m[defFDmap[redirDecl.leftFD]] = defFDmap[redirDecl.rightFD]
		} else if redirDecl.rightFD == FDUnused && redirDecl.out != "" {
			file, err := os.OpenFile(redirDecl.out, os.O_RDWR|os.O_CREATE, 0664)

			if err != nil {
				return nil, err
			}

			m[defFDmap[redirDecl.leftFD]] = file
		} else if redirDecl.rightFD == FDSupress {
			if redirDecl.out != "" {
				return nil, fmt.Errorf("Does not makes sense suppressing the output and setting an output file: >[%d=] %s", redirDecl.leftFD, redirDecl.out)
			}

			file, err := os.OpenFile("/dev/null", os.O_RDWR, 0644)

			if err != nil {
				return nil, fmt.Errorf("Cannot open /dev/null: %s", err.Error())
			}

			m[defFDmap[redirDecl.leftFD]] = file
		}

		return m, nil
	}

	return nil, fmt.Errorf("Invalid file descriptor '%d'", redirDecl.leftFD)
}

func setCmdRedir(cmd *exec.Cmd, fdMap FDMap) error {
	for k, v := range fdMap {
		switch k {
		case os.Stdin:
			cmd.Stdin = v
		case os.Stdout:
			cmd.Stdout = v
		case os.Stderr:
			cmd.Stderr = v
		default:
			return fmt.Errorf("Invalid output redirect.")
		}
	}

	return nil
}

func executeWithRedirect(cmd *exec.Cmd, redirDecls []mapDecl) error {
	var err error

	fdMap := make(FDMap)
	fdMap[os.Stdin] = os.Stdin
	fdMap[os.Stdout] = os.Stdout
	fdMap[os.Stderr] = os.Stderr

	for _, r := range redirDecls {
		fdMap, err = buildRedirect(r, fdMap)

		if err != nil {
			return err
		}
	}

	err = setCmdRedir(cmd, fdMap)

	if err != nil {
		return err
	}

	err = cmd.Start()

	if err != nil {
		return err
	}

	err = cmd.Wait()

	if err != nil {
		return err
	}

	return nil
}

func main() {
	var err error

	cmd := exec.Command("/bin/ls", "-l")

	mapDecls := make([]mapDecl, 1)
	mdecl := mapDecl{
		leftFD:  1,
		rightFD: 2,
	}

	mapDecls[0] = mdecl

	err = executeWithRedirect(cmd, mapDecls)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Command executed successfully\n")
}
