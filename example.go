package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/sefacicekli/go-http-server/core"
)

func main() {
	fmt.Print("Test sunucusu baslatiliyor...\n")
	s, err := server.Listen(8080, func(w io.Writer, req *server.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{
				StatusCode: server.StatusBadRequest,
				Message:    "Your problem is not my problem \n",
			}
		case "/myproblem":
			return &server.HandlerError{
				StatusCode: server.StatusInternalServerError,
				Message:    "Woopsie, my bad \n",
			}
		default:
			w.Write([]byte("All good, frfr\n"))
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Sunucu baslatilirken hata: %v", err)
	}
	defer s.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Sunucu kapatildi.")
}
