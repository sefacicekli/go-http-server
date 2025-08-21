package server

import (
	"fmt"
	"io"
	"log"
)

func HandleConnection(s *Server, conn io.ReadWriteCloser) {
	fmt.Print("New connection has been initialized. \n")
	r, err := RequestFromReader(conn)

	if err != nil {
		if err == io.EOF {
			fmt.Println("Connection closed by client before request was sent.")
			return 
		}
		log.Printf("An error occured while reading request: %v", err)
		return
	}

	fmt.Printf("Request Line: %s \n", r.RequestLine)
	fmt.Printf("- Method: %s \n", r.RequestLine.Method)
	fmt.Printf("- Target: %s \n", r.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s \n", r.RequestLine.HttpVersion)

	// HEADERS

	r.Headers.ForEach(func(n, v string) {
		fmt.Printf("- %s: %s \n", n ,v)
	})

	// BODY

	fmt.Printf("Body: \n")
	fmt.Printf("%s \n", r.Body)

	WriteData(s, conn, r)
}

// func disconnect(conn net.Conn) {
// 	err := conn.Close()

// 	if err == nil {
// 		fmt.Print("Connection has been disconnected from listener. \n")
// 	} else {
// 		log.Fatal("error", "An error occured while disconnecting..", err)
// 	}
// }

// func ReadData(conn *net.Conn) {
// 	content := ParseByChunk(conn)
// 	for line := range content {
// 		fmt.Printf("Read: %s \n", line)
// 	}

// }
