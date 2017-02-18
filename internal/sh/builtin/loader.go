package main

import (
	"fmt"

	"github.com/NeowayLabs/nash/sh"
)

// Fn is the contract of a built in function, that is simpler
// than the core nash Fn.
type Fn interface {
	Stringer

	ArgNames() []string
	Run() (sh.Obj, error)
	SetArgs(args []sh.Obj) error
}

func Load() map[string]Fn {

	return map[string]Fn{
		"split": newSplitFn(),
		"len":   newLenFn(),
		// FIXME: break it till you make it :-)
		//lenfn := NewLenFn(shell)
		//"len": NewLenFn(shell),
		//shell.Setvar("len", sh.NewFnObj(lenfn))

		//appendfn := NewAppendFn(shell)
		//shell.builtins["append"] = appendfn
		//shell.Setvar("append", sh.NewFnObj(appendfn))

		//splitfn := NewSplitFn(shell)
		//shell.builtins["split"] = splitfn
		//shell.Setvar("split", sh.NewFnObj(splitfn))

		//chdir := NewChdir(shell)
		//shell.builtins["chdir"] = chdir
		//shell.Setvar("chdir", sh.NewFnObj(chdir))
	}
	fmt.Println("vim-go")
}
