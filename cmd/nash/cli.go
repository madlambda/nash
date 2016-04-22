package main

import (
	"bytes"
	"fmt"

	"github.com/nemith/goline"
	"github.com/tiago4orion/nash"
)

func cli(sh *nash.Shell) error {
	var (
		err   error
		value string
	)

	gliner := goline.NewGoLine(sh)

	gliner.AddHandler(goline.CHAR_CTRLC, goline.Finish)
	gliner.AddHandler(goline.CHAR_CTRLD, goline.UserTerminated)

	var content bytes.Buffer
	var lineidx int

	for {
		if value, err = gliner.Line(); err == nil {
			fmt.Printf("\n")
			lineidx++

			content.Write([]byte(value + "\n"))

			parser := nash.NewParser(fmt.Sprintf("line %d", lineidx), string(content.Bytes()))

			tr, err := parser.Parse()

			if err != nil {
				if err.Error() == "Open '{' not closed" {
					sh.SetMultiLine(true)
					continue
				}

				fmt.Printf("ERROR: %s\n", err.Error())
				content.Reset()

				sh.SetMultiLine(false)
				continue
			}

			content.Reset()

			if value == "exit" {
				break
			}

			sh.SetMultiLine(false)

			err = sh.ExecuteTree(tr)

			if err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
			}
		} else if err == goline.UserTerminatedError {
			fmt.Println("Aborted")
			break
		} else {
			panic(err)
		}
	}

	return err
}
