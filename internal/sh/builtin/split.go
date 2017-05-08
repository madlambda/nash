package builtin

import (
	"io"
	"strings"

	"github.com/NeowayLabs/nash/errors"
	"github.com/NeowayLabs/nash/sh"
)

type (
	splitFn struct {
		content string
		sep     sh.Obj
	}
)

func newSplit() *splitFn {
	return &splitFn{}
}

func (s *splitFn) ArgNames() []string {
	return []string{"sep", "content"}
}

func (s *splitFn) Run(in io.Reader, out io.Writer, err io.Writer) ([]sh.Obj, error) {
	var output []string

	content := s.content

	switch s.sep.Type() {
	case sh.StringType:
		sep := s.sep.(*sh.StrObj).Str()
		output = strings.Split(content, sep)
	case sh.ListType:
		sepList := s.sep.(*sh.ListObj).List()
		output = splitByList(content, sepList)
	case sh.FnType:
		sepFn := s.sep.(*sh.FnObj).Fn()
		output = splitByFn(content, sepFn)
	default:
		return nil, errors.NewError("Invalid separator value: %v", s.sep)
	}

	listobjs := make([]sh.Obj, len(output))

	for i := 0; i < len(output); i++ {
		listobjs[i] = sh.NewStrObj(output[i])
	}

	return []sh.Obj{sh.NewListObj(listobjs)}, nil
}

func (s *splitFn) SetArgs(args []sh.Obj) error {
	if len(args) != 2 {
		return errors.NewError("split: expects 2 parameters")
	}

	if args[0].Type() != sh.StringType {
		return errors.NewError("split: first parameter must be a string")
	}

	content := args[0].(*sh.StrObj)

	s.content = content.Str()
	s.sep = args[1]

	return nil
}

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

		results := splitFunc.Results()

		if len(results) != 1 {
			// expects a single return fn
			return false
		}

		result := results[0]

		//FIXME: It would be cool to only accept booleans
		// since the splitter is a predicate
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
