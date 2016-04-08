package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type (
	process struct {
		pid     int
		cmdline string
	}

	namespace map[string][]process
)

var (
	ns namespace
)

func (ns namespace) PrintTree() {
	for nsname, processes := range map[string][]process(ns) {
		fmt.Printf("Namespace %s\n", nsname)

		for _, p := range processes {
			fmt.Printf("\t- %s (%d)\n", p.cmdline, p.pid)
		}
	}
}

func printFile(base string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		var pid int

		if err != nil {
			log.Print(err)
			return nil
		}

		if !info.IsDir() {
			return filepath.SkipDir
		}

		tmp := strings.Replace(path, base, "", 1)

		if tmp == "" {
			// Path == base, continue walking
			return nil
		}

		// skip remaining / in the beginning
		tmp = tmp[1:]

		fileparts := strings.Split(tmp, string(os.PathSeparator))

		tmp = tmp[1:] // skip remaining initial "/"

		if pid, err = strconv.Atoi(fileparts[0]); err != nil {
			// Linux proc thrash.. Other thing that's not a pid info
			return filepath.SkipDir
		}

		cmdpath := path + "/cmdline"
		cmdline, err := ioutil.ReadFile(cmdpath)

		if err != nil {
			log.Printf("Error: %s\n", err.Error())
			return filepath.SkipDir
		}

		if len(cmdline) == 0 {
			return filepath.SkipDir
		}

		p := process{
			pid:     pid,
			cmdline: string(cmdline[:32]),
		}

		nspaths := []string{
			path + "/ns/mnt",
			path + "/ns/ipc",
			path + "/ns/pid",
			path + "/ns/net",
			path + "/ns/uts",
			path + "/ns/user",
		}

		for _, nspath := range nspaths {
			l, err := os.Readlink(nspath)

			if err != nil {
				log.Printf("Error: %s", err.Error())
				os.Exit(1)
			}

			if ns[l] != nil {
				ns[l] = append(ns[l], p)
			} else {
				ns[l] = make([]process, 1, 4096)
				ns[l][0] = p
			}
		}

		return filepath.SkipDir
	}
}

func main() {
	log.SetFlags(log.Lshortfile)

	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Printf("Usage: %s [<optional alternative proc directory>]\n", os.Args[0])
		os.Exit(1)
	}

	dir := "/proc"

	if len(os.Args) == 2 {
		dir = os.Args[1]
	}

	ns = make(map[string][]process)

	filepath.Walk(dir, printFile(dir))

	ns.PrintTree()
}
