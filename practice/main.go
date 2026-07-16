package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	// bufio scanner
	inputData := "Line one\nLine two\nLine three"

	stream := strings.NewReader(inputData)
	scanner := bufio.NewScanner(stream)
	
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("Found Token: %v\n", line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error while reading stream", err)
	}

	// bufio reader
	delimiterData := "HEAD:SampleData#NextData"
	stream2 := strings.NewReader(delimiterData)
	reader2 := bufio.NewReader(stream2)

	preview, _ := reader2.Peek(4)
	fmt.Printf("peaked value: %s\n", preview)

	first_part, _ := reader2.ReadString('#')
	fmt.Printf("first value: %s\n", first_part)

	last_part, _ := reader2.ReadString('#')
	fmt.Printf("last value: %s\n", last_part)

	// net.listen
	address := ":4000"
	fmt.Printf("starting server at %s", address)

	listner, err := net.Listen("tcp", address)

	if err != nil {
		fmt.Println("error while starting server, ", err)
	}
	defer listner.Close()

	// block until a client connects to this address
	connection, err := listner.Accept()

	if err != nil {
		fmt.Println("error while accepting client:", err)
		return
	}
	defer connection.Close()

	fmt.Printf("client connect with address: %s, ", connection.RemoteAddr())
	fmt.Printf("Connection details: %#v\n", connection)
}
