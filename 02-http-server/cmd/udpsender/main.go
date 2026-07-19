package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	port := ":4000"
	udpAddr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		fmt.Println("error while connecting to server")
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		fmt.Println("error while dialing to udp server:", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		fmt.Println(">")

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading input: %v\n", err)
			return
		}

		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Printf("error sending UDP packet: %v\n", err)
		}
	}
}
