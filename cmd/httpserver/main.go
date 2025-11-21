package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/austin-weeks/http-from-scratch/internal/request"
	"github.com/austin-weeks/http-from-scratch/internal/response"
	"github.com/austin-weeks/http-from-scratch/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
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

func handler(w *response.Writer, r *request.Request) {
	if path, ok := strings.CutPrefix(r.RequestLine.RequestTarget, "/httpbin/"); ok {
		proxyHTTPBin(path, w)
		return
	}

	var statusCode response.StatusCode
	var body []byte
	switch r.RequestLine.RequestTarget {
	case "/video":
		if r.RequestLine.Method != "GET" {
			statusCode = response.StatusBadRequest
			break
		}
		sendVideo(w)
		return
	case "/yourproblem":
		statusCode = response.StatusBadRequest
		body = []byte(`
<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)

	case "/myproblem":
		statusCode = response.StatusInternalError
		body = []byte(`
<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
	default:
		statusCode = response.StatusOK
		body = []byte(`
<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
	}
	h := response.GetDefaultHeaders(len(body))
	h.OverwriteSet("Content-Type", "text/html")

	err := w.WriteStatusLine(statusCode)
	if err != nil {
		slog.Error("failed to write status line", "error", err, "request", r)
	}

	err = w.WriteHeaders(h)
	if err != nil {
		slog.Error("failed to write headers", "error", err, "request", r)
	}

	_, err = w.WriteBody(body)
	if err != nil {
		slog.Error("failed to write body", "error", err, "request", r)
	}
}
