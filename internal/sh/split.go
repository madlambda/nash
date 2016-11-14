package sh

import (
	"io"
	"strings"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	SplitFn struct {
		stdin          io.Reader
		stdout, stderr io.Writer

		done    chan struct{}
		err     error
		results sh.Obj

		content string
		sep     sh.Obj
	}
)

func NewSplitFn(env *Shell) *SplitFn {
	return &SplitFn{
		stdin:  env.stdin,
		stdout: env.stdout,
		stderr: env.stderr,
	}
}

func (splitfn *SplitFn) Name() string {
	return "split"
}

func (splitfn *SplitFn) ArgNames() []string {
	return []string{
		"sep",
		"content",
	}
}

func (splitfn *SplitFn) run() error {
	var output []string

	content := splitfn.content

	switch splitfn.sep.Type() {
	case sh.StringType:
		sep := splitfn.sep.(*sh.StrObj).Str()
		output = strings.Split(content, sep)
	case sh.ListType:
		sepList := splitfn.sep.(*sh.ListObj).List()
		output = splitByList(content, sepList)
	case sh.FnType:
		sepFn := splitfn.sep.(*sh.FnObj).Fn()
		output = splitByFn(content, sepFn)
	default:
		return errors.NewError("Invalid separator value: %v", splitfn.sep)
	}

	listobjs := make([]sh.Obj, len(output))

	for i := 0; i < len(output); i++ {
		listobjs[i] = sh.NewStrObj(output[i])
	}

	splitfn.results = sh.NewListObj(listobjs)
	return nil
}

func (splitfn *SplitFn) Start() error {
	splitfn.done = make(chan struct{})

	go func() {
		splitfn.err = splitfn.run()
		splitfn.done <- struct{}{}
	}()

	return nil
}

func (splitfn *SplitFn) Wait() error {
	<-splitfn.done
	return splitfn.err
}

func (splitfn *SplitFn) Results() sh.Obj {
	return splitfn.results
}

func (splitfn *SplitFn) SetArgs(args []sh.Obj) error {
	if len(args) != 2 {
		return errors.NewError("splitfn expects 2 arguments")
	}

	if args[0].Type() != sh.StringType {
		return errors.NewError("content must be of type string")
	}

	content := args[0].(*sh.StrObj)

	splitfn.content = content.Str()
	splitfn.sep = args[1]

	return nil
}

func (splitfn *SplitFn) SetEnviron(env []string) {
	// do nothing
}

func (splitfn *SplitFn) SetStdin(r io.Reader)  { splitfn.stdin = r }
func (splitfn *SplitFn) SetStderr(w io.Writer) { splitfn.stderr = w }
func (splitfn *SplitFn) SetStdout(w io.Writer) { splitfn.stdout = w }
func (splitfn *SplitFn) StdoutPipe() (io.ReadCloser, error) {
	return nil, errors.NewError("splitfn doesn't works with pipes")
}
func (splitfn *SplitFn) Stdin() io.Reader  { return splitfn.stdin }
func (splitfn *SplitFn) Stdout() io.Writer { return splitfn.stdout }
func (splitfn *SplitFn) Stderr() io.Writer { return splitfn.stderr }

func (splitfn *SplitFn) String() string { return "<builtin fn split>" }

func splitByList(content string, delims []sh.Obj) []string {
	return strings.FieldsFunc(content, func(r rune) bool {
		for _, delim := range delims {
			if delim.Type() != sh.StringType {
				continue
			}

			objstr := delim.(*sh.StrObj)

			if len(objstr.Str()) > 0 && rune(objstr.Str()[0]) == r {
				return true
			}
		}

		return false
	})
}

func splitByFn(content string, splitFunc sh.Fn) []string {
	return strings.FieldsFunc(content, func(r rune) bool {
		arg := sh.NewStrObj(string(r))
		splitFunc.SetArgs([]sh.Obj{arg})
		err := splitFunc.Start()

		if err != nil {
			return false
		}

		err = splitFunc.Wait()

		if err != nil {
			return false
		}

		result := splitFunc.Results()

		if result.Type() != sh.StringType {
			return false
		}

		status := result.(*sh.StrObj)

		if status.Str() == "0" {
			return true
		}

		return false
	})
}
