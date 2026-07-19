package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// 1. start-line
	reqLine, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("error while reading start-line")
		return
	}
	reqLine = strings.TrimRight(reqLine, "\r\n")
	parts := strings.Split(reqLine, " ")

	if len(parts) != 3 {
		fmt.Println("incorrect http start-line")
		return
	}

	method, path, protocol := parts[0], parts[1], parts[2]
	fmt.Printf("method: %s, path: %s, protocol: %s\n", method, path, protocol)

	// 2. headers
	headers := make(map[string]string)
	var contentLength int64 = 0

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("error while reading header: ", err)
			break
		}

		line = strings.TrimRight(line, "\r\n")
		// only \r\n in line means, end of headers
		if line == "" {
			break
		}

		lineSplit := strings.Split(line, ":")
		if len(lineSplit) != 2 {
			continue
		}

		key := strings.TrimSpace(lineSplit[0])
		val := strings.TrimSpace(lineSplit[1])
		headers[key] = val

		if strings.ToLower(key) == "content-length" {
			if cl, err := strconv.ParseInt(val, 10, 64); err == nil {
				contentLength = cl
			}
		}
	}

	fmt.Println("")
	fmt.Println("header:")
	for key, val := range headers {
		fmt.Printf("%s: %s\n", key, val)
	}
	fmt.Println("")

	// 3. body
	var bodyBytes []byte
	if contentLength > 0 {
		limitedReader := io.LimitReader(reader, contentLength)
		bodyBytes, err = io.ReadAll(limitedReader)

		if err != nil {
			fmt.Printf("error reading body: %v\n", err)
			return
		}

		fmt.Printf("Body: %s\n", string(bodyBytes))
	}

	response := "HTTP/1.1 200 OK\r\nContent-Length: 11\r\nContent-Type: text/plain\r\n\r\nHello World"
	conn.Write([]byte(response))
}

func main() {
	port := ":4000"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("error while starting server")
		return
	}
	fmt.Printf("server started at: %s\n", port)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("client connection failed with error: %s\n", err)
			break
		}

		go handleConnection(conn)
	}
}
