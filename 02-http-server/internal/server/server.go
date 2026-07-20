package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"mytcpserver/internal/request"
	"mytcpserver/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	listener net.Listener
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
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
		handler:  handler,
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

func sendResponse(conn net.Conn, body string, statusCode response.StatusCode) error {
	contentLength := len(body) + len("\r\n")

	writer := io.Writer(conn)

	err := response.WriteStatusLine(writer, statusCode)
	if err != nil {
		return nil
	}

	headers := response.GetDefaultHeaders(contentLength)
	err = response.WriteHeaders(writer, headers)
	if err != nil {
		return err
	}

	err = response.WriteMessageBody(writer, body)
	return err
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		sendResponse(conn, err.Error(), response.StatBadRequest)
		return
	}
	req.PrintRequestLine()

	body := bytes.NewBuffer([]byte{})
	handleError := s.handler(body, req)
	if handleError != nil {
		sendResponse(conn, handleError.Message, handleError.StatusCode)
		return
	}

	sendResponse(conn, body.String(), response.StatusOK)
}

func (s *Server) Close() error {
	s.cancel()
	err := s.listener.Close()

	s.wg.Wait()
	return err
}
