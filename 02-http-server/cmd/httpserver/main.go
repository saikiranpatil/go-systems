package main

import (
	"log"
	"mytcpserver/internal/request"
	"mytcpserver/internal/response"
	"mytcpserver/internal/server"
	"os"
	"os/signal"
	"syscall"
)

const port = 4000

const (
	HTML400 string = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

	HTML500 string = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Server Error</h1>
    <p>Something went wrong on our end.</p>
  </body>
</html>`

	HTML200 string = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
)

func main() {
	var handler server.Handler = func(res *response.Response, req *request.Request) *server.HandlerError {
		res.SetHeader("Content-Type", "text/html")

		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{
				StatusCode: response.StatBadRequest,
				Message:    HTML400,
			}

		case "/myproblem":
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    HTML500,
			}

		case "/video":
			res.Headers.Replace("Content-Type", "video/mp4")

			b, _ := os.ReadFile("./assets/videoplayback.mp4")
			res.WriteMessageBody(b)
			return nil

		case "/image":
			res.Headers.Replace("Content-Type", "image/png")

			b, _ := os.ReadFile("./assets/avatar.png")
			res.WriteMessageBody(b)
			return nil

		default:
			res.SetHeader("X-Sample-Header-Value", "Sample Value")
			res.WriteStatusLine(response.StatusOK)
			res.WriteMessageBody([]byte(HTML200))
			return nil
		}
	}

	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	// 3. Keep the process alive safely
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
