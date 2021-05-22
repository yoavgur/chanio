package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/yoavgur/chanio"
)

func startClient(wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("Connecting to " + connType + " server " + connHost + ":" + connPort)

	conn, err := net.Dial(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println("Error connecting:", err.Error())
		os.Exit(1)
	}

	stdin := chanio.NewReader(os.Stdin)
	reader := chanio.NewReader(conn)
	writer := chanio.NewWriter(conn)

	for {
		select {
		case input, ok := <-stdin.Read():
			if !ok {
				return
			}
			writer.Write() <- input

		case resp, ok := <-reader.Read():
			if !ok {
				return
			}
			fmt.Println("Server:", strings.Trim(string(resp), "\n"))
		}
	}
}
