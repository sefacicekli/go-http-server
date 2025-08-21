package server

import (
	"fmt"
	"log"
	"net"
	"io"
	// "strconv"
)

type HandlerError struct {
	StatusCode StatusCode
	Message    string
}

type Handler func(w io.Writer, req *Request) *HandlerError 

type Server struct{
	closed bool
	handler Handler
}

func runServer(s *Server, listener net.Listener) {
	for {
		conn, error := listener.Accept()
		if s.closed {
			return
		}
		if error != nil {
			log.Fatal("error", "An error occured while listening tcp: ", error)
		}

		go HandleConnection(s, conn)
	}
}

func Listen(port uint16, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	server := &Server{
		closed: false,
		handler: handler,
	}
	go runServer(server, listener)
	if err != nil {
		log.Fatal("error", "An error occured while listening tcp: ", err)
	}

	return server, nil

}

func (s *Server) Close() error {
	s.closed = true
	return nil
}