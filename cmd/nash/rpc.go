package main

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/NeowayLabs/nash"
)

func serveConn(sh *nash.Shell, conn net.Conn) {
	var data [1024]byte

	for {

		n, err := conn.Read(data[:])

		if err != nil {
			if err == io.EOF {
				return
			}

			fmt.Printf("Failed to read data: %s", err.Error())
			return
		}

		if string(data[0:n]) == "quit" {
			return
		}

		err = sh.ExecuteString("-nashd-", string(data[0:n]))

		if err != nil {
			fmt.Printf("nashd: %s\n", err.Error())

			_, err = conn.Write([]byte("1"))

			if err != nil {
				fmt.Printf("Failed to send command status.\n")
				return
			}
		} else {
			_, err = conn.Write([]byte("0"))

			if err != nil {
				fmt.Printf("Failed to send command status.\n")
				return
			}
		}
	}
}

func startNashd(sh *nash.Shell, socketPath string) {
	os.Remove(socketPath)

	addr := &net.UnixAddr{
		Net:  "unix",
		Name: socketPath,
	}

	listener, err := net.ListenUnix("unix", addr)

	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		return
	}

	// Accept only one connection
	conn, err := listener.AcceptUnix()

	if err != nil {
		fmt.Printf("ERROR: %v", err.Error())
	}

	serveConn(sh, conn)
	listener.Close()
}
