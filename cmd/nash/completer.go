package main

import (
	"os"
	"path/filepath"
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
	goNext := false
	var lineCompleter readline.PrefixCompleterInterface

	for _, child := range c.prefixCompleter.GetChildren() {
		childName := child.GetName()
		if len(line) >= len(childName) {
			if runes.HasPrefix(line, childName) {
				if len(line) == len(childName) {
					newLine = append(newLine, []rune{' '})
				} else {
					newLine = append(newLine, childName)
				}
				offset = len(childName)
				lineCompleter = child
				goNext = true
			}
		} else {
			if runes.HasPrefix(childName, line) {
				newLine = append(newLine, childName[len(line):])
				offset = len(line)
				lineCompleter = child
			}
		}
	}

	newLine, offset = c.completePaths(line, newLine, offset)

	if len(newLine) != 1 {
		return
	}

	tmpLine := make([]rune, 0, len(line))
	for i := offset; i < len(line); i++ {
		if line[i] == ' ' {
			continue
		}

		tmpLine = append(tmpLine, line[i:]...)
		return lineCompleter.Do(tmpLine, len(tmpLine))
	}

	if goNext {
		return lineCompleter.Do(nil, 0)
	}
	return
}

func (c *Completer) completePaths(line []rune, oldline [][]rune, oldoffset int) (newLine [][]rune, offset int) {
	newLine = oldline
	offset = oldoffset

	paths, ok := c.sh.GetEnv("PATH")

	if !ok {
		return
	}

	for _, pathval := range paths {
		for _, base := range strings.Split(pathval, ":") {
			filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}

				if info.IsDir() && path != base {
					return filepath.SkipDir
				}

				tmp := []rune(strings.Replace(path, base, "", 1))

				if len(tmp) == 0 {
					// Path == base, continue walking
					return nil
				}

				// skip remaining / in the beginning
				if tmp[0] == '/' {
					tmp = tmp[1:]
				}

				if len(line) >= len(tmp) {
					if runes.HasPrefix(line, tmp) {
						if len(line) == len(tmp) {
							newLine = append(newLine, []rune{' '})
						} else {
							newLine = append(newLine, tmp)
						}

						offset = len(tmp)
					}
				} else {
					if runes.HasPrefix(tmp, line) {
						newLine = append(newLine, tmp[len(line):])
						offset = len(line)
					}
				}

				return nil
			})
		}
	}

	return
}
