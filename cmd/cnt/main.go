package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tiago4orion/cnt"
)

var (
	debug     int
	file      string
	rforkAddr string
)

func init() {
	flag.IntVar(&debug, "debug", 0, "debug level")
	flag.StringVar(&rforkAddr, "rforkAddr", "", "rfork unix file")
}

func main() {
	var err error

	flag.Parse()

	if rforkAddr != "" {
		startRpcServer(rforkAddr, debug)
	} else if file == "" {
		err = cli(debug)
	} else {
		err = cnt.Execute(file, debug)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}
}
