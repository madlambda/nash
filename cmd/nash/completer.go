package main

import (
	"io/ioutil"
	"strings"

	"github.com/NeowayLabs/nash"
	"github.com/chzyer/readline"
)

var runes = readline.Runes{}

type Completer struct {
	sh              *nash.Shell
	prefixCompleter readline.PrefixCompleterInterface
}

func NewCompleter(sh *nash.Shell, p ...readline.PrefixCompleterInterface) *Completer {
	return &Completer{
		prefixCompleter: readline.NewPrefixCompleter(p...),
		sh:              sh,
	}
}

func (c *Completer) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	line = runes.TrimSpaceLeft(line[:pos])

	if len(line) >= 1 && (line[0] == '/' || line[0] == '.') {
		return completeFile(line, pos, "")
	} else if len(line) == 0 {
		return completeFile([]rune{'/'}, 0, "/")
	}

	pathVar, ok := c.sh.GetEnv("PATH")

	if !ok {
		return
	}

	path := make([]string, 0, 256)

	for _, pathVal := range pathVar {
		pathparts := strings.Split(pathVal, ":")
		if len(pathparts) == 1 {
			path = append(path, pathparts[0])
		} else {
			for _, p := range pathparts {
				path = append(path, p)
			}
		}
	}

	return completeInPathList(path, line, pos)
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
