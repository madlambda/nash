package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/NeowayLabs/nash"
)

func main() {
	var (
		file io.ReadCloser
		err  error
	)

	if len(os.Args) <= 1 {
		file = os.Stdin
	} else {
		fname := os.Args[1]
		file, err = os.Open(fname)

		if err != nil {
			log.Fatal("[ERROR] " + err.Error())
		}
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

	fmt.Printf("%s\n", ast)
}
