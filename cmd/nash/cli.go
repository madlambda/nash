package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/NeowayLabs/nash"
	"github.com/NeowayLabs/nash/ast"
	"github.com/NeowayLabs/nash/parser"
	"github.com/NeowayLabs/nash/sh"
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

func execFn(shell *nash.Shell, fn sh.Fn) {
	fn.SetStdin(shell.Stdin())
	fn.SetStdout(shell.Stdout())
	fn.SetStderr(shell.Stderr())

	if err := fn.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %s\n", fn.Name(), err.Error())
		return
	}

	if err := fn.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %s\n", fn.Name(), err.Error())
		return
	}
}

func cli(shell *nash.Shell) error {
	var err error

	historyFile := shell.DotDir() + "/history"

	cfg := readline.Config{
		Prompt:          shell.Prompt(),
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	}

	term, err := readline.NewTerminal(&cfg)

	if err != nil {
		return err
	}

	op := term.Readline()
	rline := &readline.Instance{
		Config:    &cfg,
		Terminal:  term,
		Operation: op,
	}

	defer rline.Close()

	completer := NewCompleter(op, term, shell)

	cfg.AutoComplete = completer

	if lineMode, ok := shell.Getvar("LINEMODE"); ok {
		if lineStr, ok := lineMode.(*sh.StrObj); ok && lineStr.Str() == "vim" {
			rline.SetVimMode(true)
		} else {
			rline.SetVimMode(false)
		}
	}

	return docli(shell, rline)
}

func docli(shell *nash.Shell, rline *readline.Instance) error {
	var (
		content    bytes.Buffer
		lineidx    int
		line       string
		parse      *parser.Parser
		tr         *ast.Tree
		err        error
		unfinished bool
		prompt     string
	)

	for {
		if fn, ok := shell.GetFn("nash_repl_before"); ok && !unfinished {
			execFn(shell, fn)
		}

		if !unfinished {
			prompt = shell.Prompt()
		}

		rline.SetPrompt(prompt)

		line, err = rline.Readline()

		if err == readline.ErrInterrupt {
			goto cont
		} else if err == io.EOF {
			err = nil
			break
		}

		lineidx++

		line = strings.TrimSpace(line)

		// handle special cli commands

		switch {
		case strings.HasPrefix(line, "set mode "):
			switch line[9:] {
			case "vi":
				rline.SetVimMode(true)
			case "emacs":
				rline.SetVimMode(false)
			default:
				fmt.Printf("invalid mode: %s\n", line[9:])
			}

			goto cont
		case line == "mode":
			if rline.IsVimMode() {
				fmt.Printf("Current mode: vim\n")
			} else {
				fmt.Printf("Current mode: emacs\n")
			}

			goto cont

		case line == "exit":
			break
		}

		content.Write([]byte(line + "\n"))

		parse = parser.NewParser(fmt.Sprintf("<stdin line %d>", lineidx), string(content.Bytes()))

		line = string(content.Bytes())

		tr, err = parse.Parse()

		if err != nil {
			if interrupted, ok := err.(Interrupted); ok && interrupted.Interrupted() {
				content.Reset()
				goto cont
			} else if errBlock, ok := err.(BlockNotFinished); ok && errBlock.Unfinished() {
				prompt = ">>> "
				unfinished = true
				goto cont
			}

			fmt.Printf("ERROR: %s\n", err.Error())

			content.Reset()
			goto cont
		}

		unfinished = false
		content.Reset()

		_, err = shell.ExecuteTree(tr)

		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
		}

	cont:
		if fn, ok := shell.GetFn("nash_repl_after"); ok && !unfinished {
			var status sh.Obj
			var ok bool

			if status, ok = shell.Getvar("status"); !ok {
				status = sh.NewStrObj("")
			}

			err = fn.SetArgs([]sh.Obj{sh.NewStrObj(line), status})

			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
			} else {
				execFn(shell, fn)
			}
		}

		rline.SetPrompt(prompt)
	}

	return nil
}
