package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/tiago4orion/cnt"
)

type RforkRpc error

func (r *RforkRpc) ExecuteNode(node cnt.Node, reply *error) error {
	fmt.Printf("Executing node: %v\n", node)
	return nil
}

func startRpcServer(socketPath string, debug int) {
	os.Remove(socketPath)

	go func() {
		myRpc := new(RforkRpc)

		if err := rpc.Register(RforkRpc); err != nil {
			log.Fatal(err)
		}
		addr := &net.UnixAddr{Net: "unix", Name: socketPath}
		listener, err := net.ListenUnix("unix", addr)
		if err != nil {
			log.Fatal(err)
		}
		for {
			conn, err := listener.AcceptUnix()
			if err != nil {
				log.Fatal(err)
			}
			rpc.ServeConn(conn)
		}
	}()

	time.Sleep(2 * time.Second)
	log.Println("server exiting")
}
