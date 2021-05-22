package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

func main() {
	fmt.Println("Hello")

	// command := exec.Command("bash", "-c", "tcpdump -nni any")
	// command := exec.Command("bash", "-c", "while true; do echo hi; sleep 1; done")
	// command := exec.Command("bash", "-c", "tail -f /tmp/fucker")

	// stdout, err := command.StdoutPipe()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// if command.Start() != nil {
	// 	fmt.Println("fuck")
	// 	return
	// }

	// command2 := exec.Command("bash", "-c", "cat > /tmp/fucker")

	// stdin, err := command2.StdinPipe()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// stdout2, err := command2.StdoutPipe()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// if command2.Start() != nil {
	// 	fmt.Println("fuck")
	// 	return
	// }

	r, w, _ := os.Pipe()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	writer := NewWriter(w)

	readers := NewReaderContext(r, ctx)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			writer.Write() <- []byte{0x70, 0x6f, 0x6f, 0x70}
			err := writer.GetError()
			if err != nil {
				fmt.Println("FUCKER", err)
				return
			}
			time.Sleep(time.Second)
		}
	}()

	a := true
	for a {
		select {
		// case writer.Write() <- []byte{0x70, 0x6f, 0x6f, 0x70}:
		// 	// fmt.Println("Wrote")
		// 	time.Sleep(time.Second)
		case buf, ok := <-readers.Read():
			if !ok {
				fmt.Println("Done reading")
				a = false
			}
			fmt.Println("Reader:", string(buf))
			// default:
			// 	fmt.Println("can't write")
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

	// for out := range readers.Read() {
	// 	fmt.Print(string(out))
	// }

	fmt.Println("\nDone:", readers.GetError())
	fmt.Println("\nDone:", writer.GetError())
	wg.Wait()
}
