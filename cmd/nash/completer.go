package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/madlambda/nash"
	"github.com/madlambda/nash/readline"
	"github.com/madlambda/nash/sh"
)

var runes = readline.Runes{}

type Completer struct {
	op   *readline.Operation
	term *readline.Terminal
	sh   *nash.Shell
}

func NewCompleter(op *readline.Operation, term *readline.Terminal, sh *nash.Shell) *Completer {
	return &Completer{op, term, sh}
}

func (c *Completer) Do(line []rune, pos int) ([][]rune, int) {
	var (
		newLine [][]rune
		offset  int
		lineArg = sh.NewStrObj(string(line))
		posArg  = sh.NewStrObj(strconv.Itoa(pos))
	)

	defer c.op.Refresh()
	defer c.term.PauseRead(false)

	fnDef, err := c.sh.GetFn("nash_complete")
	if err != nil {
		// no complete available
		return [][]rune{[]rune{'\t'}}, offset
	}

	nashFunc := fnDef.Build()
	err = nashFunc.SetArgs([]sh.Obj{lineArg, posArg})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to autocomplete: %s\n", err.Error())
		return newLine, offset
	}

	nashFunc.SetStdin(c.sh.Stdin())
	nashFunc.SetStdout(c.sh.Stdout())
	nashFunc.SetStderr(c.sh.Stderr())

	if err = nashFunc.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to autocomplete: %s\n", err.Error())
		return newLine, offset
	}

	if err = nashFunc.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to autocomplete: %s\n", err.Error())
		return newLine, offset
	}

	ret := nashFunc.Results()

	if len(ret) != 1 || ret[0].Type() != sh.ListType {
		fmt.Fprintf(os.Stderr, "ignoring autocomplete value: %v\n", ret)
		return newLine, offset
	}

	retval := ret[0]

	retlist := retval.(*sh.ListObj)

	if len(retlist.List()) != 2 {
		return newLine, pos
	}

	newline := retlist.List()[0]
	newpos := retlist.List()[1]

	if newline.Type() != sh.StringType || newpos.Type() != sh.StringType {
		fmt.Fprintf(os.Stderr, "ignoring autocomplete value: (%s) (%s)\n", newline, newpos)
		return newLine, offset
	}

	objline := newline.(*sh.StrObj)
	objpos := newpos.(*sh.StrObj)

	newoffset, err := strconv.Atoi(objpos.Str())

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to autocomplete: %s\n", err.Error())
		return newLine, offset
	}

	newLine = append(newLine, []rune(objline.Str()))

	return newLine, newoffset
}
