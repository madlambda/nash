package main

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

var runes = readline.Runes{}

type Completer struct {
	prefixCompleter readline.PrefixCompleterInterface
}

func NewCompleter(p ...readline.PrefixCompleterInterface) *Completer {
	return &Completer{
		prefixCompleter: readline.NewPrefixCompleter(p...),
	}
}

func (c *Completer) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	var local bool

	line = runes.TrimSpaceLeft(line[:pos])

	if len(line) == 0 {
		return
	}

	completeStr := line

	for i := pos - 1; i >= 0; i-- {
		if line[i] == ' ' {
			completeStr = completeStr[i+1 : pos]
			local = true
			break
		}
	}

	if len(completeStr) > 0 {
		for i := 0; i < len(completeStr); i++ {
			if completeStr[i] == ' ' {
				completeStr = completeStr[:i]
				break
			}
		}
	}

	if (len(completeStr) > 0 && (completeStr[0] == '/' || completeStr[0] == '.')) || local {
		return completeFile(line, completeStr, pos, "")
	}

	pathVal := os.Getenv("PATH")
	path := make([]string, 0, 256)

	pathparts := strings.Split(pathVal, ":")
	if len(pathparts) == 1 {
		path = append(path, pathparts[0])
	} else {
		for _, p := range pathparts {
			path = append(path, p)
		}
	}

	return completeInPathList(path, line, completeStr, pos)
}

func completeInPath(path string, line, complete []rune, offset int) ([][]rune, int, bool) {
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

func completeInPathList(pathList []string, line, complete []rune, offset int) ([][]rune, int) {
	var newOffset int

	newLine := make([][]rune, 0, 256)

	for _, path := range pathList {
		tmpNewLine, tmpOffset, found := completeInPath(path, line, complete, offset)

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

func completeCurrentPath(line, complete []rune, offset int) ([][]rune, int) {
	lineStr := string(complete[2:])
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

		if len(cmpStr) >= len(string(complete)) &&
			strings.HasPrefix(cmpStr, string(complete)) {

			if len(cmpStr) == len(string(complete)) {
				newLine = append(newLine, []rune{' '})

				offset = len(cmpStr)
			} else {
				newLine = append(newLine, []rune(cmpStr[len(string(complete)):]))
			}
		}
	}

	return newLine, offset
}

func completeAbsolutePath(line, complete []rune, offset int, prefix string) ([][]rune, int) {
	lineStr := string(complete[1:]) // ignore first '/'
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

		if len(cmpStr) >= len(string(complete)) && strings.HasPrefix(cmpStr, string(complete)) {
			if len(cmpStr) == len(string(complete)) {
				newLine = append(newLine, []rune{' '})

				offset = len(cmpStr)
			} else {
				newLine = append(newLine, []rune(prefix+cmpStr[len(string(complete)):]))
			}
		}
	}

	return newLine, offset
}

func completeFile(line, complete []rune, offset int, prefix string) ([][]rune, int) {
	llen := len(complete)

	if llen >= 1 {
		if complete[0] != '/' {
			if llen >= 2 && complete[0] == '.' && complete[1] == '/' {
				return completeCurrentPath(line, complete, offset)
			} else {
				return completeCurrentPath(line, []rune("./"+string(complete)), offset)
			}
		}

		return completeAbsolutePath(line, complete, offset, prefix)

	} else {
		return completeFile(line, []rune{'.', '/'}, 1, prefix)
	}
}
