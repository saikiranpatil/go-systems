package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer f.Close()

		str := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}

			message := string(data[:n])

			if idx := strings.Index(message, "\n"); idx != -1 {
				str += message[:idx]
				message = message[idx+1 : n]

				ch <- str
				str = ""
			}

			str += message
		}

		if len(str) != 0 {
			ch <- str
		}
	}()

	return ch
}

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println("error while reading file")
	}

	ch := getLinesChannel(file)

	for messages:= range ch {
		fmt.Printf("read: %s\n", messages)
	}
}
