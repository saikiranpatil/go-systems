package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string

const (
	StateInit  parserState = "init"
	StateDone  parserState = "done"
	StateError parserState = "error"
)

type Request struct {
	RequestLine RequestLine
	state       parserState
}

// constants
var crlf = "\r\n"

// errors
var ErrorBadStartLine = fmt.Errorf("bad start line")
var ErrorMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")

func getInitialRequest() *Request {
	return &Request{
		state: StateInit,
	}
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, []byte(crlf))
	if idx == -1 {
		return nil, 0, ErrorBadStartLine
	}

	read := idx + len(crlf)

	parts := strings.Split(string(b[:idx]), " ")
	if len(parts) != 3 {
		return nil, 0, ErrorMalformedRequestLine
	}

	httpParts := strings.Split(string(parts[2]), "/")
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return nil, 0, ErrorUnsupportedHttpVersion
	}

	rl := &RequestLine{
		Method:        parts[0],
		HttpVersion:   httpParts[1],
		RequestTarget: parts[1],
	}

	return rl, read, nil
}

func (r *Request) shouldParseRequestLine() bool {
	return r.state == StateDone || r.state == StateInit
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		switch r.state {
		case StateError:
			return 0, ErrorMalformedRequestLine

		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				if err == ErrorBadStartLine {
					// no crlf yet, not an error
					break outer
				}

				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateDone
		case StateDone:
			break outer
		default:
			panic("unreachable request state")
		}
	}

	return read, nil
}

func (r *Request) PrintRequestLine() {
	requestLine := r.RequestLine
	if requestLine.Method == "" {
		return
	}

	fmt.Printf(
		"Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
		requestLine.Method,
		requestLine.RequestTarget,
		requestLine.HttpVersion,
	)
}

func RequestFromReader(reader io.ReadCloser) (*Request, error) {
	defer func() {
		reader.Close()
		fmt.Printf("client closed!\n")
	}()
	r := getInitialRequest()

	buf := make([]byte, 1024)
	bufLen := 0

	for r.shouldParseRequestLine() {
		if bufLen == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			if err == io.EOF {
				if r.state != StateDone {
					return nil, fmt.Errorf("incomplete request, in state %s", r.state)
				}
				break
			}
			return nil, err
		}

		bufLen += n
		readN, err := r.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return r, nil
}
