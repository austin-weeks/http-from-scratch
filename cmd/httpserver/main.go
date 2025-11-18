package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/austin-weeks/http-from-scratch/internal/request"
	"github.com/austin-weeks/http-from-scratch/internal/response"
	"github.com/austin-weeks/http-from-scratch/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    "Your problem is not my problem\n",
			}
		case "/myproblem":
			return &server.HandlerError{
				StatusCode: response.StatusInternalError,
				Message:    "Woopsie, my bad\n",
			}
		default:
			_, err := w.Write([]byte("All good, frfr\n"))
			if err != nil {
				return &server.HandlerError{
					StatusCode: response.StatusInternalError,
					Message:    err.Error(),
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close() // nolint
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
