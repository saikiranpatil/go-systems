package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"mytcpserver/internal/request"
)

type Server struct {
	listener net.Listener
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

func Serve(port int) (*Server, error) {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to port %d: %w", port, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	srv := &Server{
		listener: listener,
		ctx:      ctx,
		cancel:   cancel,
	}

	go srv.listenLoop()

	return srv, nil
}

func (s *Server) listenLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

func sendResponse(conn net.Conn) error {
	body := "Hello World!, this is sample message"
	contentLength := len(body)

	_, err := fmt.Fprintf(conn,
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/plain\r\n"+
			"Content-Length: %d\r\n"+
			"\r\n"+
			"%s",
		contentLength, body,
	)

	return err
}
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Printf("Error parsing request: %v\n", err)
		return
	}
	req.PrintRequestLine()
	if err := sendResponse(conn); err != nil {
		fmt.Printf("Error sending response: %v\n", err)
	}
}

func (s *Server) Close() error {
	s.cancel()
	err := s.listener.Close()

	s.wg.Wait()
	return err
}
