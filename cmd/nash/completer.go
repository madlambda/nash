package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/NeowayLabs/nash"
	"github.com/NeowayLabs/nash/sh"
	"github.com/chzyer/readline"
)

var runes = readline.Runes{}

type Completer struct {
	term *readline.Terminal
	sh   *nash.Shell
}

func NewCompleter(term *readline.Terminal, sh *nash.Shell) *Completer {
	return &Completer{term, sh}
}

func (c *Completer) Do(line []rune, pos int) ([][]rune, int) {
	var (
		newLine [][]rune
		offset  int
		lineArg = sh.NewStrObj(string(line))
		posArg  = sh.NewStrObj(strconv.Itoa(pos))
	)

	defer c.term.PauseRead(false)

	nashFunc, ok := c.sh.GetFn("nash_complete")

	if !ok {
		return newLine, offset
	}

	err := nashFunc.SetArgs([]sh.Obj{lineArg, posArg})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to autocomplete: %s", err.Error())
		return newLine, offset
	}

	nashFunc.SetStdin(c.sh.Stdin())
	nashFunc.SetStdout(c.sh.Stdout())
	nashFunc.SetStderr(c.sh.Stderr())

	if err = nashFunc.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to autocomplete: %s", err.Error())
		return newLine, offset
	}

	if err = nashFunc.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to autocomplete: %s", err.Error())
		return newLine, offset
	}

	ret := nashFunc.Results()

	if ret == nil || ret.Type() != sh.ListType {
		fmt.Fprintf(os.Stderr, "ignoring autocomplete value: %v\n", ret)
		return newLine, offset
	}

	retlist := ret.(*sh.ListObj)

	if len(retlist.List()) != 2 {
		fmt.Fprintf(os.Stderr, "ignoring autocomplete value: %v (%d)\n", retlist, len(retlist.List()))
		return newLine, offset
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
		fmt.Fprintf(os.Stderr, "Failed to autocomplete: %s", err.Error())
		return newLine, offset
	}

	newLine = append(newLine, []rune(objline.Str()))

	return newLine, newoffset
}
