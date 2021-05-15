package main

import (
	"fmt"
	"os/exec"
)

func main() {
	fmt.Println("Hello")

	command := exec.Command("find", "/")

	stdout, err := command.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	if command.Start() != nil {
		fmt.Println("fuck")
		return
	}

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()

	reader := NewReader(stdout)

	output := reader.Read()

	for out := range output {
		fmt.Print(string(out))
	}

	fmt.Println("\nDone: ", reader.GetError())
}
