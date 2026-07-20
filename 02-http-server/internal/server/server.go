package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"mytcpserver/internal/request"
	"mytcpserver/internal/response"
)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(res *response.Response, req *request.Request) *HandlerError

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
	response := response.NewResponse(conn)

	err := response.WriteStatusLine(statusCode)
	if err != nil {
		return nil
	}

	headers := response.Headers
	headers.Replace("Content-Type", "text/plain")
	err = response.WriteHeaders()
	if err != nil {
		return err
	}

	err = response.WriteMessageBody([]byte(body))
	return err
}
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	res := response.NewResponse(conn)
	if err != nil {
		res.WriteStatusLine(response.StatBadRequest)
		res.WriteMessageBody([]byte(err.Error()))
		return
	}
	req.PrintRequestLine()

	handleError := s.handler(res, req)
	if handleError != nil {
		res.WriteStatusLine(handleError.StatusCode)
		res.WriteMessageBody([]byte(handleError.Message))
		return
	}
}

func (s *Server) Close() error {
	s.cancel()
	err := s.listener.Close()

	s.wg.Wait()
	return err
}
