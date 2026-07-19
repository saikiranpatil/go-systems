package main

import (
	"fmt"
	"mytcpserver/internal/request"
	"net"
)

func main() {
	port := ":4000"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("error while sarting server")
		return
	}
	defer listener.Close()
	fmt.Printf("server started at %s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error while connecting client")
			continue
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println("error while reading request: ", err)
			continue
		}

		r.PrintRequestLine()
	}
}
