package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/NeowayLabs/nash"
	"github.com/NeowayLabs/nash/sh"
	"github.com/chzyer/readline"
)

var runes = readline.Runes{}

type Completer struct {
	sh *nash.Shell
}

func NewCompleter(sh *nash.Shell) *Completer {
	return &Completer{sh}
}

func (c *Completer) Do(line []rune, pos int) ([][]rune, int) {
	var (
		newLine [][]rune
		offset  int
		lineArg = sh.NewStrObj(string(line))
		posArg  = sh.NewStrObj(strconv.Itoa(pos))
	)

	nashFunc, ok := c.sh.GetFn("nash_complete")

	if !ok {
		return newLine, offset
	}

	err := nashFunc.SetArgs([]sh.Obj{lineArg, posArg})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to autocomplete: %s", err.Error())
		return newLine, offset
	}

	if err = nashFunc.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to autocomplete: %s", err.Error())
		return newLine, offset
	}

	if err = nashFunc.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to autocomplete: %s", err.Error())
		return newLine, offset
	}

	ret := nashFunc.Results()

	if ret.Type() != sh.ListType {
		fmt.Fprintf(os.Stderr, "ignoring autocomplete value: %v\n", ret)
		return newLine, offset
	}

	retlist := ret.(*sh.ListObj)

	if len(retlist.List()) != 2 {
		fmt.Fprintf(os.Stderr, "ignoring autocomplete value: %v\n", retlist)
		return newLine, offset
	}

	newline := retlist.List()[0]
	newpos := retlist.List()[1]

	if newline.Type() != sh.StringType || newpos.Type() != sh.StringType {
		fmt.Fprintf(os.Stderr, "ignoring autocomplete value: %v\n", retlist)
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

	fmt.Printf("newLine = %v, pos = %d\n", newLine, newoffset)

	time.Sleep(time.Second * 5)
	return newLine, newoffset
}

func completeInPath(path string, line []rune, offset int) ([][]rune, int, bool) {
	var found bool

	newLine := make([][]rune, 0, 256)

	if len(line) == 0 {
		return newLine, offset, found
	}

	files, err := ioutil.ReadDir(path)

	if err != nil {
		return newLine, offset, found
	}

	for _, file := range files {
		fname := file.Name()

		if len(string(line)) <= len(fname) && strings.HasPrefix(fname, string(line)) {
			if len(string(line)) == len(fname) {
				newLine = append(newLine, []rune{' '})
				offset = len(fname)
				found = true
				break
			} else {
				newLine = append(newLine, []rune(fname[len(string(line)):]))
			}
		}
	}

	return newLine, offset, found
}

func completeInPathList(pathList []string, line []rune, offset int) ([][]rune, int) {
	var newOffset int

	newLine := make([][]rune, 0, 256)

	for _, path := range pathList {
		tmpNewLine, tmpOffset, found := completeInPath(path, line, offset)

		if len(tmpNewLine) > 0 {
			newLine = append(newLine, tmpNewLine...)
			newOffset = tmpOffset
		}

		if found {
			break
		}
	}

	return newLine, newOffset
}

func completeCurrentPath(line []rune, offset int) ([][]rune, int) {
	lineStr := string(line[2:])
	dirParts := strings.Split(lineStr, "/")
	directory := "./" + strings.Join(dirParts[0:len(dirParts)-1], "/")

	newLine := make([][]rune, 0, 256)

	files, err := ioutil.ReadDir(directory)

	if err != nil {
		return newLine, offset
	}

	for _, file := range files {
		var cmpStr string

		fname := file.Name()

		if directory[len(directory)-1] == '/' {
			cmpStr = directory + fname
		} else {
			cmpStr = directory + "/" + fname
		}

		if len(cmpStr) >= len(string(line)) &&
			strings.HasPrefix(cmpStr, string(line)) {

			if len(cmpStr) == len(string(line)) {
				newLine = append(newLine, []rune{' '})

				offset = len(cmpStr)
			} else {
				newLine = append(newLine, []rune(cmpStr[len(string(line)):]))
			}
		}
	}

	return newLine, offset
}

func completeAbsolutePath(line []rune, offset int, prefix string) ([][]rune, int) {
	lineStr := string(line[1:]) // ignore first '/'
	dirParts := strings.Split(lineStr, "/")
	directory := "/" + strings.Join(dirParts[0:len(dirParts)-1], "/")

	newLine := make([][]rune, 0, 256)

	files, err := ioutil.ReadDir(directory)

	if err != nil {
		return newLine, offset
	}

	for _, file := range files {
		var cmpStr string

		fname := file.Name()

		if directory[len(directory)-1] == '/' {
			cmpStr = directory + fname
		} else {
			cmpStr = directory + "/" + fname
		}

		if len(cmpStr) >= len(string(line)) && strings.HasPrefix(cmpStr, string(line)) {
			if len(cmpStr) == len(string(line)) {
				newLine = append(newLine, []rune{' '})

				offset = len(cmpStr)
			} else {
				newLine = append(newLine, []rune(prefix+cmpStr[len(string(line)):]))
			}
		}
	}

	return newLine, offset
}

func completeFile(line []rune, offset int, prefix string) ([][]rune, int) {
	llen := len(line)

	if llen >= 1 {
		if line[0] != '/' {
			if llen >= 2 && line[0] == '.' && line[1] == '/' {
				return completeCurrentPath(line, offset)
			}

			return [][]rune{}, offset
		}

		return completeAbsolutePath(line, offset, prefix)

	} else {
		return completeFile([]rune{'/'}, 1, prefix)
	}
}
