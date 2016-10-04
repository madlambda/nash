package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/NeowayLabs/nash"
	"github.com/NeowayLabs/nash/parser"
	"github.com/chzyer/readline"
)

type (
	Interrupted interface {
		Interrupted() bool
	}

	Ignored interface {
		Ignore() bool
	}

	BlockNotFinished interface {
		Unfinished() bool
	}
)

var completers = []readline.PrefixCompleterInterface{}

func cli(sh *nash.Shell) error {
	var (
		err error
	)

	historyFile := sh.DotDir() + "/history"

	cfg := readline.Config{
		Prompt:          sh.Prompt(),
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	}

	term, err := readline.NewTerminal(&cfg)

	if err != nil {
		return err
	}

	op := term.Readline()
	l := &readline.Instance{
		Config:    &cfg,
		Terminal:  term,
		Operation: op,
	}

	completer := NewCompleter(op, term, sh)

	cfg.AutoComplete = completer

	defer l.Close()

	//	log.SetOutput(l.Stderr())

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

		parse := parser.NewParser(fmt.Sprintf("line %d", lineidx), string(content.Bytes()))

		tr, err := parse.Parse()

		if err != nil {
			if interrupted, ok := err.(Interrupted); ok && interrupted.Interrupted() {
				l.SetPrompt(sh.Prompt())
				continue
			}

			if errBlock, ok := err.(BlockNotFinished); ok && errBlock.Unfinished() {
				l.SetPrompt(">>> ")
				continue
			}

			fmt.Printf("ERROR: %s\n", err.Error())
			content.Reset()
			l.SetPrompt(sh.Prompt())
			continue
		}

		content.Reset()

		_, err = sh.ExecuteTree(tr)

		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
		}

		l.SetPrompt(sh.Prompt())
	}

	return err
}
