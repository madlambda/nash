package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func serveConn(conn net.Conn) {
	fmt.Printf("New connection: %v", conn)

	var data [1024]byte

	for {

		n, err := conn.Read(data[:])

		if err != nil {
			fmt.Printf("Failed to read data: %s", err.Error())
			return
		}

		fmt.Printf("Read '%d' bytes\n", n)
		fmt.Printf("Value: %q\n", string(data[0:n]))

		time.Sleep(1 * time.Second)
	}
}

func startRcd(socketPath string, debug bool) {
	os.Remove(socketPath)

	var wg sync.WaitGroup

	fmt.Printf("Starting server: %s\n", socketPath)

	wg.Add(1)

	go func() {
		addr := &net.UnixAddr{Net: "unix", Name: socketPath}

		listener, err := net.ListenUnix("unix", addr)

		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			return
		}

		for {
			conn, err := listener.AcceptUnix()

			if err != nil {
				fmt.Printf("ERROR: %v", err.Error())
			}

			serveConn(conn)
		}
	}()

	wg.Wait()
}
