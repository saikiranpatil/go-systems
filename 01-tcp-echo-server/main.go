package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

func main() {
	port := ":8000"
	listner, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("error while starting server: ", err)
		return
	}

	defer listner.Close()
	fmt.Printf("server started at %s", port)

	for {
		conn, err := listner.Accept()
		if err != nil {
			fmt.Println("error while listening client: ", err)
			continue
		}

		go handleConnect(conn)
	}
}

func handleConnect(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("error while reading string: ", err)
			}
			break
		}

		fmt.Printf("Received from %s: %s", conn.RemoteAddr(), message)

		_, err = conn.Write([]byte("message: "+message))
		if err != nil {
			fmt.Printf("Error writing to %s: %v", conn.RemoteAddr(), err)
			break
		}
	}
}
