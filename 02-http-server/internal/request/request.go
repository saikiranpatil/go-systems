package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"mytcpserver/internal/headers"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type parserState string

const (
	StateInit           parserState = "init"
	StateDone           parserState = "done"
	StateParsingHeaders parserState = "parsingHeaders"
	StateParsingBody    parserState = "parsingBody"
)

type Request struct {
	RequestLine RequestLine
	state       parserState
	Headers     headers.Headers
	Body        []byte
}

// constants
var crlf = "\r\n"

// errors
var ErrorBadStartLine = fmt.Errorf("bad start line")
var ErrorMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrorUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrorInvalidContentLength = fmt.Errorf("invalid content length")

func getInitialRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
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

func parseFeildLine(b []byte) {

}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				if err == ErrorBadStartLine {
					// no crlf yet, not an error
					break outer
				}

				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = StateParsingHeaders

		case StateParsingHeaders:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer // need more data from the network before we can continue
			}

			read += n
			if done {
				r.state = StateParsingBody
			}

		case StateParsingBody:
			contentLengthStr, found := r.Headers.Get("Content-Length")
			// content-length not found,
			// meaning no body present
			// hance continue
			if !found {
				r.state = StateDone
				continue
			}

			contentLength, err := strconv.Atoi(contentLengthStr)
			if err != nil {
				return 0, err
			}

			if len(r.Body) == contentLength {
				r.state = StateDone
				continue
			}

			neededBytes := contentLength - len(r.Body)
			if neededBytes < 0 {
				return 0, ErrorInvalidContentLength
			}

			availableBytes := len(data) - read
			if availableBytes == 0 {
				break outer
			}

			availableBytes = min(availableBytes, neededBytes)
			r.Body = append(r.Body, data[read:read+availableBytes]...)
			read += availableBytes

			if len(r.Body) == contentLength {
				r.state = StateDone
			} else {
				break outer // Need more data from the network
			}

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

	// print request line
	fmt.Printf(
		"Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
		requestLine.Method,
		requestLine.RequestTarget,
		requestLine.HttpVersion,
	)

	// print field line
	fmt.Println("Headers:")
	for key, value := range r.Headers {
		fmt.Printf("- %s: %s\n", key, value)
	}

	// request body
	fmt.Println("Body:")
	fmt.Println(string(r.Body))
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := getInitialRequest()

	buf := make([]byte, 1024)
	bufLen := 0

	for r.state != StateDone {
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
