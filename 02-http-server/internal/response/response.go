package response

import (
	"errors"
	"fmt"
	"io"
	"mytcpserver/internal/headers"
	"net"
	"strconv"
)

var crlf = "\r\n"

var (
	ErrStatusLineAlreadySent = errors.New("status line must be written first")
	ErrHeadersAlreadySent    = errors.New("headers have already been sent down the socket")
	ErrBodyAlreadySent       = errors.New("body has already been flushed and connection finalized")
	ErrInvalidCallSequence   = errors.New("method execution violates structural lifecycle rules")
	ErrCannotSetHeaders      = errors.New("cannot set headers after they are sent to the client")
	ErrInvalidTrailerCall    = errors.New("cannot set trailer before status line or headers or body")
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatBadRequest            StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type responseState string

const (
	stateStatusLine responseState = "responseLine"
	stateHeaders    responseState = "headers"
	stateBody       responseState = "body"
	stateDone       responseState = "done"
)

type Response struct {
	StatusCode StatusCode
	Headers    headers.Headers
	Body       []byte
	state      responseState
	writer     io.Writer
}

func NewResponse(conn net.Conn) *Response {
	writer := io.Writer(conn)

	h := headers.NewHeaders()
	h.Set("Content-Type", "text/plain")

	response := &Response{
		state:   stateStatusLine,
		writer:  writer,
		Headers: h,
	}

	return response
}

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

func (r *Response) WriteStatusLine(statusCode StatusCode) error {
	if r.state != stateStatusLine {
		return ErrStatusLineAlreadySent
	}

	r.StatusCode = statusCode
	reasonPhrase := getReasonPhrase(statusCode)
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)

	_, err := r.writer.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	r.state = stateHeaders
	return nil
}

func (r *Response) WriteHeaders() error {
	if r.state == stateStatusLine {
		if err := r.WriteStatusLine(StatusOK); err != nil {
			return err
		}
	}
	if r.state != stateHeaders {
		return ErrHeadersAlreadySent
	}

	for key, value := range r.Headers {
		fieldLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := r.writer.Write([]byte(fieldLine))
		if err != nil {
			return err
		}
	}

	_, err := r.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	r.state = stateBody
	return nil
}

func (r *Response) WriteMessageBody(body []byte) error {
	if r.state == stateDone {
		return ErrBodyAlreadySent
	}

	if r.state == stateStatusLine || r.state == stateHeaders {
		r.Headers.Replace("Content-Length", strconv.Itoa(len(body)))
		if err := r.WriteHeaders(); err != nil {
			return err
		}
	}

	r.Body = body
	_, err := r.writer.Write(r.Body)
	if err != nil {
		return err
	}

	r.state = stateDone
	return nil
}

func (r *Response) SetHeader(key, value string) error {
	if r.state == stateBody || r.state == stateDone {
		return ErrCannotSetHeaders
	}

	r.Headers.Set(key, value)
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func (r *Response) WriteChunkedBody(buf []byte, bufLen int) (int, error) {
	if r.state == stateDone {
		return 0, ErrBodyAlreadySent
	}
	if r.state == stateStatusLine || r.state == stateHeaders {
		r.Headers.Delete("Content-Length")
		r.Headers.Replace("Transfer-Encoding", "chunked")
		if err := r.WriteHeaders(); err != nil {
			return 0, err
		}
	}

	totalWritten := 0

	n, err := fmt.Fprintf(r.writer, "%x\r\n", bufLen)
	if err != nil {
		return 0, err
	}
	totalWritten += n

	if bufLen > 0 {
		n, err := r.writer.Write(buf[:bufLen])
		if err != nil {
			return totalWritten, err
		}
		totalWritten += n
	}

	// 3. Close the chunk block
	n, err = r.WriteCrlf()
	if err != nil {
		return totalWritten, err
	}
	totalWritten += n

	return totalWritten, nil
}

// writes 0/r/n to the writer
func (r *Response) WriteChunkedBodyDone() (int, error) {
	n, err := r.WriteChunkedBody([]byte{}, 0)
	if err != nil {
		return 0, err
	}

	r.state = stateDone
	return n, nil
}

func (r *Response) WriteCrlf() (int, error) {
	return r.writer.Write([]byte("\r\n"))
}

func (r *Response) WriteTrailers(trailers headers.Headers) error {
	if r.state != stateDone {
		return ErrInvalidTrailerCall
	}

	for key, value := range trailers {
		trailer := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := r.writer.Write([]byte(trailer))
		if err != nil {
			return err
		}
	}

	_, err := r.writer.Write([]byte("\r\n"))
	return err
}
