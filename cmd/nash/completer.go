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
	const op = "Completer.Do"

	defer c.op.Refresh()
	defer c.term.PauseRead(false)

	fnDef, err := c.sh.GetFn("nash_complete")
	if err != nil {
		c.Log(op, "skipping autocompletion")
		return [][]rune{[]rune{'\t'}}, 0
	}

	nashFunc := fnDef.Build()
	lineArg := sh.NewStrObj(string(line))
	posArg := sh.NewStrObj(strconv.Itoa(pos))
	err = nashFunc.SetArgs([]sh.Obj{lineArg, posArg})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s:error setting args on autocomplete function:%v\n", op, err)
		return nil, 0
	}

	nashFunc.SetStdin(c.sh.Stdin())
	nashFunc.SetStdout(c.sh.Stdout())
	nashFunc.SetStderr(c.sh.Stderr())

	if err = nashFunc.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s:error starting autocomplete function:%v\n", op, err)
		return nil, 0
	}

	if err = nashFunc.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "%s:error waiting for autocomplete function:%v\n", op, err)
		return nil, 0
	}

	ret := nashFunc.Results()

	if len(ret) != 1 || ret[0].Type() != sh.ListType {
		fmt.Fprintf(os.Stderr, "%s:ignoring unexpected autocomplete func return (expected list):%+v\n", op, ret)
		return nil, 0
	}

	retlist := ret[0].(*sh.ListObj)

	if len(retlist.List()) != 2 {
		c.Log(op, "no results from autocomplete")
		return nil, pos
	}

	newline := retlist.List()[0]
	newpos := retlist.List()[1]

	if newline.Type() != sh.StringType || newpos.Type() != sh.StringType {
		fmt.Fprintf(os.Stderr, "%s:ignoring autocomplete value:(%s) (%s)\n", op, newline, newpos)
		return nil, 0
	}

	objline := newline.(*sh.StrObj)
	objpos := newpos.(*sh.StrObj)

	c.Log(op, "autocomplete result:line %q:pos %q", objline, objpos)

	offset, err := strconv.Atoi(objpos.Str())

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s:autocomplete func returned non number position:%v\n", op, err)
		return nil, 0
	}

	c.Log(op, "success:line %q:offset %d", objline, offset)
	return [][]rune{[]rune(objline.Str())}, offset
}

func (c *Completer) Log(op string, format string, args ...interface{}) {
	c.sh.Log(op+":"+format, args...)
}
