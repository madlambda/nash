package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/NeowayLabs/nash"
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
		log.Fatal("[ERROR] " + err.Error())
	}

	defer file.Close()

	content, err := ioutil.ReadAll(file)

	if err != nil {
		log.Printf("[ERROR] " + err.Error())
		return
	}

	parser := nash.NewParser("nashfmt", string(content))

	ast, err := parser.Parse()

	if err != nil {
		log.Printf("[ERROR] " + err.Error())
		return
	}

	if !overwrite {
		fmt.Printf("%s\n", ast.String())
	} else if ast.String() != string(content) {
		file.Close()

		err = ioutil.WriteFile(fname, []byte(fmt.Sprintf("%s\n", ast.String())), 0666)

		if err != nil {
			log.Printf("[ERROR] " + err.Error())
			return
		}
	}
}
