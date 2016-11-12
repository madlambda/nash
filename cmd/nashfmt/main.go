package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/NeowayLabs/nash/parser"
)

var (
	overwrite bool
)

func init() {
	flag.BoolVar(&overwrite, "w", false, "overwrite file")
}

func main() {
	var (
		file io.ReadCloser
		err  error
	)

	flag.Parse()

	if len(flag.Args()) <= 0 {
		flag.PrintDefaults()
		return
	}

	fname := flag.Args()[0]

	file, err = os.Open(fname)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	content, err := ioutil.ReadAll(file)

	if err != nil {
		file.Close()
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	parser := parser.NewParser("nashfmt", string(content))

	ast, err := parser.Parse()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		file.Close()
		os.Exit(1)
	}

	file.Close()

	if !overwrite {
		fmt.Printf("%s\n", ast.String())
		return
	}

	if ast.String() != string(content) {
		err = ioutil.WriteFile(fname, []byte(fmt.Sprintf("%s\n", ast.String())), 0666)

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
			return
		}
	}
}
