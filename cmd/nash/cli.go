package main

// [27 91 51 49 109 206 187 27 91 48 109 32

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/NeowayLabs/nash"
	"github.com/chzyer/readline"
)

type (
	Interrupted interface {
		Interrupted() bool
	}

	Ignored interface {
		Ignore() bool
	}
)

var completers = []readline.PrefixCompleterInterface{
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItem("rfork",
		readline.PcItem("c"),
		readline.PcItem("upmnis"),
		readline.PcItem("upmis"),
	),
}

func cli(sh *nash.Shell) error {
	var (
		err error
	)

	historyFile := sh.DotDir() + "/history"

	for envName, _ := range sh.Environ() {
		completers = append(completers, readline.PcItem(envName))
	}

	completer := NewCompleter(sh, completers...)

	l, err := readline.NewEx(&readline.Config{
		Prompt:          sh.Prompt(),
		HistoryFile:     historyFile,
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})

	if err != nil {
		panic(err)
	}

	defer l.Close()

	log.SetOutput(l.Stderr())

	var content bytes.Buffer
	var lineidx int
	var line string

	for {
		line, err = l.Readline()

		if err == readline.ErrInterrupt {
			continue
		} else if err == io.EOF {
			err = nil
			break
		}

		lineidx++

		line = strings.TrimSpace(line)

		// handle special cli commands

		switch {
		case strings.HasPrefix(line, "set mode "):
			switch line[8:] {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			default:
				fmt.Printf("invalid mode: %s\n", line[8:])
			}

			continue
		case line == "mode":
			if l.IsVimMode() {
				fmt.Printf("Current mode: vim\n")
			} else {
				fmt.Printf("Current mode: emacs\n")
			}

			continue

		case line == "exit":
			break
		}

		content.Write([]byte(line + "\n"))

		parser := nash.NewParser(fmt.Sprintf("line %d", lineidx), string(content.Bytes()))

		tr, err := parser.Parse()

		if err != nil {
			if interrupted, ok := err.(Interrupted); ok && interrupted.Interrupted() {
				l.SetPrompt(sh.Prompt())
				continue
			}

			if err.Error() == "Open '{' not closed" {
				l.SetPrompt(">>> ")
				continue
			}

			fmt.Printf("ERROR: %s\n", err.Error())
			content.Reset()
			l.SetPrompt(sh.Prompt())
			continue
		}

		content.Reset()

		err = sh.ExecuteTree(tr)

		if err != nil {
			if errIgnored, ok := err.(Ignored); ok && errIgnored.Ignore() {
				l.SetPrompt(sh.Prompt())
				continue
			}

			fmt.Printf("ERROR: %s\n", err.Error())
		}

		l.SetPrompt(sh.Prompt())
	}

	return err
}
