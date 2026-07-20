package response

import (
	"fmt"
	"io"
	"mytcpserver/internal/headers"
	"strconv"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatBadRequest            StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func getReasonPhrase(statusCode StatusCode) string {
	switch statusCode {
	case StatusOK:
		return "OK"
	case StatBadRequest:
		return "Bad Request"
	case StatusInternalServerError:
		return "Internal Server Error"
	default:
		return ""
	}
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reasonPhrase := getReasonPhrase(statusCode)
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)

	_, err := w.Write([]byte(statusLine))
	return err
}

func WriteMessageBody(w io.Writer, body string) error {
	messageBody:= fmt.Sprintf("\r\n%s\r\n", body)
	_, err := w.Write([]byte(messageBody))
	return err
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		fieldLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Write([]byte(fieldLine))
		if err != nil {
			return err
		}
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	// set default headers
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}