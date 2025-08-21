package server

import (
	"bytes"
	"fmt"
	"io"
	"github.com/sefacicekli/go-http-server/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func GetDefaultHeaders(contentLen int) *headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeaders(w io.Writer, h *headers.Headers) error {
	b := []byte{}
	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)
		fmt.Printf("write headers: %s \n", b)
	})
	b = fmt.Append(b, "\r\n")
	_, err := w.Write(b)

	return err
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var statusLine []byte
	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return fmt.Errorf("unrecognized status code just happened")
	}

	_, err := w.Write(statusLine)
	return err
}

func WriteData(s *Server, conn io.ReadWriteCloser, r *Request) {
	defer conn.Close()

	headers := GetDefaultHeaders(0)
	fmt.Printf("headers debug: %s", headers)
	defer fmt.Printf("headers: %s", headers)

	writer := bytes.NewBuffer([]byte{})
	handlerError := s.handler(writer, r)

	var body []byte = nil
	var status StatusCode = StatusOK
	if handlerError != nil {
		status = handlerError.StatusCode
		body = []byte(handlerError.Message)
	} else {
		body = writer.Bytes()
	}

	headers.Replace("Content-Length", fmt.Sprintf("%d", len(body)))
	WriteStatusLine(conn, status)
	WriteHeaders(conn, headers)
	conn.Write(body)
}
