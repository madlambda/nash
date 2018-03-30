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

func execFn(shell *nash.Shell, fnDef sh.FnDef, args []sh.Obj) {
	fn := fnDef.Build()
	err := fn.SetArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %s\n", fnDef.Name(), err.Error())
	}
	fn.SetStdin(shell.Stdin())
	fn.SetStdout(shell.Stdout())
	fn.SetStderr(shell.Stderr())

	if err := fn.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %s\n", fnDef.Name(), err.Error())
		return
	}

	if err := fn.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "%s failed: %s\n", fnDef.Name(), err.Error())
		return
	}
}

func importInitFile(shell *nash.Shell, initFile string) (bool, error) {
	if d, err := os.Stat(initFile); err == nil {
		if m := d.Mode(); !m.IsDir() {
			err := shell.ExecuteString("init",
				fmt.Sprintf("import %q", initFile))
			if err != nil {
				return false, fmt.Errorf("Failed to evaluate '%s': %s", initFile, err.Error())
			}
			return true, nil
		}
	}
	return false, nil
}

func loadInit(shell *nash.Shell) error {
	
	if noInit {
		return nil
	}

	initFiles := []string{
		shell.DotDir() + "/init",
		shell.DotDir() + "/init.sh",
	}

	for _, init := range initFiles {
		imported, err := importInitFile(shell, init)
		if err != nil {
			return err
		}
		if imported {
			break
		}
	}

	return nil
}

func cli(shell *nash.Shell) error {

	shell.SetInteractive(true)

	if err := loadInit(shell); err != nil {
		fmt.Fprintf(os.Stderr, "error loading init file:\n%s\n", err)
	}

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
		if fnDef, err := shell.GetFn("nash_repl_before"); err == nil && !unfinished {
			execFn(shell, fnDef, nil)
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
		if fnDef, err := shell.GetFn("nash_repl_after"); err == nil && !unfinished {
			var status sh.Obj
			var ok bool

			if status, ok = shell.Getvar("status"); !ok {
				status = sh.NewStrObj("")
			}

			execFn(shell, fnDef, []sh.Obj{sh.NewStrObj(line), status})
		}

		rline.SetPrompt(prompt)
	}

	return nil
}
