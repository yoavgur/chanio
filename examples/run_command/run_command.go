package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/yoavgur/chanio"
)

func main() {
	// Print hi in loop forever
	command := exec.Command("bash", "-c", "while true; do echo hi; sleep 1; done")

	// Get stdout reader
	stdout, err := command.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get stderr reader
	stderr, err := command.StderrPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Run command
	err = command.Start()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer command.Wait()

	// Generate context for reading from readers
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Create a multi-reader to read from stderr and stdout simultaneously
	// Pass in our context, which will allow us to break from reading after
	// 10 seconds without worrying that we'll be blocked
	reader := chanio.NewMultiReaderContext(ctx, stdout, stderr)

	// Read until read channel is closed
	for chunk := range reader.Read() {
		fmt.Print(string(chunk))
	}

	// We timed out, kill our created process
	command.Process.Kill()

	// Print any errors which occured
	fmt.Printf("Finished reading with errors: %v\n", reader.GetErrors())
}
