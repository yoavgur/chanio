package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/yoavgur/chanio"
)

func startServer(wg *sync.WaitGroup) {
	defer wg.Done()

	l, err := net.Listen(connType, connHost+":"+connPort)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error connecting:", err.Error())
			return
		}

		fmt.Println("Client " + c.RemoteAddr().String() + " connected.")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
		defer cancel()
		go handleConnection(c, ctx)
	}
}

func handleConnection(conn net.Conn, ctx context.Context) {
	reader := chanio.NewReaderContext(conn, ctx)
	writer := chanio.NewWriterContext(conn, ctx)

	for buf := range reader.Read() {
		fmt.Println("Client:", strings.Trim(string(buf), "\n"))
		writer.Write() <- buf
	}
}
