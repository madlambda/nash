package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"
	"github.com/tiago4orion/nash"
)

var (
	history_fn = filepath.Join(os.TempDir(), ".nash_history")
	names      = []string{"rfork"}
)

func cli(sh *nash.Shell) error {
	var (
		err   error
		value string
	)

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(false)

	line.SetCompleter(func(line string) (c []string) {
		for _, n := range names {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}

		return
	})

	if f, err := os.Open(history_fn); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	var content bytes.Buffer
	var lineidx int

	for {
		if value, err = line.Prompt("nash> "); err == nil {
			lineidx++

			content.Write([]byte(value + "\n"))

			parser := nash.NewParser(fmt.Sprintf("line %d", lineidx), string(content.Bytes()))

			tr, err := parser.Parse()

			if err != nil {
				if err.Error() == "Open '{' not closed" {
					continue
				}

				fmt.Printf("ERROR: %s\n", err.Error())
				content.Reset()
				continue
			} else {
				line.AppendHistory(string(content.Bytes()))
			}

			content.Reset()

			if value == "exit" {
				break
			}

			err = sh.ExecuteTree(tr)

			if err != nil {
				fmt.Printf("ERROR: %s\n", err.Error())
			}
		} else if err == liner.ErrPromptAborted {
			log.Print("Aborted")
			break
		} else {
			if err.Error() == "EOF" {
				err = nil
				break
			}

			log.Print("Error reading line: ", err)
		}

		if err != nil {
			fmt.Printf("%s\n", err.Error())
		}
	}

	if f, err := os.Create(history_fn); err != nil {
		log.Print("Error writing history file: ", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}

	return err
}
