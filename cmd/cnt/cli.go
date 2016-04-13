package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"
	"github.com/tiago4orion/cnt"
)

var (
	history_fn = filepath.Join(os.TempDir(), ".cnt_history")
	names      = []string{"rfork"}
)

func cli(debugval bool) error {
	var (
		err   error
		value string
	)

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

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

	for {
		if value, err = line.Prompt("cnt> "); err == nil {
			line.AppendHistory(value)

			if value == "exit" {
				break
			}

			err = cnt.ExecuteString("<input>", value, debugval)
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
