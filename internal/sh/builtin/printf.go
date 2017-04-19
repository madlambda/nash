package builtin

import (
	"fmt"
	"io"

	"github.com/NeowayLabs/nash/sh"
)

type (
	printfFn struct {
		fmt  string
		args []sh.Obj
	}
)

func newPrintf() *printfFn {
	return &printfFn{}
}

func (s *printfFn) ArgNames() []string {
	return []string{"fmt", "args"}
}

func (s *printfFn) Run(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) ([]sh.Obj, error) {
	fmt.Fprintf(stdout, "helloworld")
	return nil, nil
}

func (s *printfFn) SetArgs(args []sh.Obj) error {
	//if len(args) != 1 {
	//return errors.NewError("splitfn expects 2 arguments")
	//}

	//if args[0].Type() != sh.StringType {
	//return errors.NewError("content must be of type string")
	//}
	//content := args[0].(*sh.StrObj)
	//s.content = content.Str()
	//s.sep = args[1]
	return nil
}
