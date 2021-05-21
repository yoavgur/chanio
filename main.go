package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

func main() {
	fmt.Println("Hello")

	// command := exec.Command("bash", "-c", "tcpdump -nni any")
	// command := exec.Command("bash", "-c", "while true; do echo hi; sleep 1; done")
	command := exec.Command("bash", "-c", "tail -f /tmp/fucker")

	stdout, err := command.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	if command.Start() != nil {
		fmt.Println("fuck")
		return
	}

	command2 := exec.Command("bash", "-c", "cat > /tmp/fucker")

	stdin, err := command2.StdinPipe()
	if err != nil {
		fmt.Println(err)
		return
	}

	// stdout2, err := command2.StdoutPipe()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	if command2.Start() != nil {
		fmt.Println("fuck")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	writer := NewWriterContext(stdin, ctx)

	readers := NewMultiReader(1024, ctx, stdout)

	for {
		select {
		case writer.Write() <- []byte{0x70, 0x6f, 0x6f, 0x70}:
			// time.Sleep(time.Second)
		case buf, ok := <-readers.Read():
			if !ok {
				fmt.Println("Done reading")
				return
			}
			fmt.Println("Reader:", string(buf))
			// default:
			// fmt.Println("can't write")
		}
	}

	// for {
	// 	buf := make([]byte, 1024)
	// 	nb_read, err := io.ReadAtLeast(readers, buf, 1)
	// 	if err != nil {
	// 		break
	// 	}
	// 	fmt.Print(string(buf[:nb_read]))
	// }

	// reader := NewReaderContext(stdout, ctx)

	// output := reader.Read()

	// fuck := true
	// for fuck {
	// 	select {
	// 	case out, ok := <-output:
	// 		if !ok {
	// 			fuck = false
	// 		} else {
	// 			fmt.Print(string(out))
	// 		}

	// 	case <-time.After(time.Second):
	// 		cancel()
	// 	}
	// }

	for out := range readers.Read() {
		fmt.Print(string(out))
	}

	fmt.Println("\nDone:", readers.GetErrors())
}
