package server

import (
	"bytes"
	"strconv"
	"fmt"
	"io"
	"github.com/sefacicekli/go-http-server/headers"
)

type HTTP_METHOD string
type parserState string

const (
	GET    HTTP_METHOD = "GET"
	POST   HTTP_METHOD = "POST"
	PUT    HTTP_METHOD = "PUT"
	DELETE HTTP_METHOD = "DELETE"
	PATCH  HTTP_METHOD = "PATCH"
)

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateDone    parserState = "done"
	StateError   parserState = "error"
	StateBody    parserState = "body"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        HTTP_METHOD
}

var ErrMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrMalformedHeaders = fmt.Errorf("malformed headers")
var ErrUnsupportedHTTPVersion = fmt.Errorf("unsupported http version")
var ErrRequestInErrorState = fmt.Errorf("request in error state")

func (rl *RequestLine) Validate() bool {
	return rl.HttpVersion == "HTTP/1.1"
}

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        string
	state       parserState
}

func (r *Request) hasBody() bool {
	// When doing chunk encoding update this method
	length := getInt(r.Headers, "content-length", 0)
	return length > 0

}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}

		switch r.state {
		case StateError:
			return 0, ErrRequestInErrorState
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			r.state = StateHeaders
		case StateHeaders:

			n, done, err := r.Headers.Parse(currentData)

			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

		case StateBody:
			length := getInt(r.Headers, "content-length", 0)

			if length <= 0 {
				panic("chunked not implemented")
			}

			remaining := min(length-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining
			if len(r.Body) == length {
				r.state = StateDone
			}

		case StateDone:
			break outer
		default:
			panic("Somehow we have programmed terrible")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func getInt(header *headers.Headers, name string, defaultValue int) int {
	valueStr, exists := header.Get(name)

	if !exists {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)

	if err != nil {
		return defaultValue
	}

	return value
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body:    "",
	}
}

var SEPERATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPERATOR)

	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPERATOR)

	parts := bytes.Split(startLine, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, ErrMalformedRequestLine
	}

	rl := &RequestLine{
		HttpVersion:   string(parts[2]),
		RequestTarget: string(parts[1]),
		Method:        HTTP_METHOD(parts[0]),
	}

	if !rl.Validate() {
		return nil, 0, ErrUnsupportedHTTPVersion
	}

	return rl, read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// NOTE: BUFFER COULD GET OVERRUN
	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		// TODO: WHAT TO DO HERE?
		if err != nil {
			return nil, err
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])

		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}

// func parseHeaders(h string) ([]string, error) {
// 	idx := strings.Index(h, SEPERATOR)

// 	if idx == -1 {
// 		return nil, ErrMalformedHeaders
// 	}

// }
