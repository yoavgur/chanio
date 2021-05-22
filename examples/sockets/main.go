package main

import "sync"

const (
	connHost = "localhost"
	connPort = "9090"
	connType = "tcp"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go startServer(&wg)

	wg.Add(1)
	go startClient(&wg)

	wg.Wait()
}
