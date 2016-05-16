package main

import (
	"fmt"
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
			}
		} else {
			if runes.HasPrefix(childName, line) {
				newLine = append(newLine, childName[len(line):])
				offset = len(line)
			}
		}
	}

	newLine, offset = c.completePaths(line, newLine, offset)

	return
}

func (c *Completer) completeCurrentPath(line []rune, oldline [][]rune, oldoffset int) (newLine [][]rune, offset int) {
	filepath.Walk(".", func(dirpath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if dirpath == "." {
			return nil
		}

		path := []rune("./" + dirpath)

		//		fmt.Printf("PAth=%s\n", string(path))

		if len(line) >= len(path) {
			if runes.HasPrefix(line, path) {
				if len(line) == len(path) {
					newLine = append(newLine, []rune{' '})
				} else {
					newLine = append(newLine, path)
				}

				offset = len(path)
			}
		} else {
			if runes.HasPrefix(path, line) {
				newLine = append(newLine, path[len(line):])
				offset = len(line)
			}
		}

		if info.IsDir() {
			return filepath.SkipDir
		}

		return nil
	})

	fmt.Printf("Returning: %q\n", newLine)

	return
}

func (c *Completer) completePaths(line []rune, oldline [][]rune, oldoffset int) (newLine [][]rune, offset int) {
	newLine = oldline
	offset = oldoffset

	if runes.HasPrefix(line, []rune("./")) {
		return c.completeCurrentPath(line, oldline, oldoffset)
	}

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
