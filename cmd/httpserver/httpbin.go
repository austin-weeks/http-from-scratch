package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/austin-weeks/http-from-scratch/internal/headers"
	"github.com/austin-weeks/http-from-scratch/internal/response"
)

func proxyHTTPBin(path string, w *response.Writer) {
	path = fmt.Sprintf("https://httpbin.org/%s", path)
	r, err := http.Get(path)
	if err != nil {
		slog.Error("failed to fetch from httpbin", "error", err, "path", path)
		return
	}
	defer r.Body.Close() // nolint

	err = w.WriteStatusLine(response.StatusOK)
	if err != nil {
		slog.Error("failed to write status line", "error", err, "path", path)
		return
	}
	h := headers.NewHeaders()
	h.Set("Connection", "close")
	h.Set("Content-Type", "application/json")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailers", "X-Content-SHA256, X-Content-Length")

	err = w.WriteHeaders(h)
	if err != nil {
		slog.Error("failed to write headers", "error", err, "path", path)
		return
	}

	var body []byte
	buf := make([]byte, 32)
	for {
		n, err := r.Body.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				slog.Error("failed to read httpbin response", "error", err, "path", path)
			}
			break
		}
		body = append(body, buf[:n]...)
		_, err = w.WriteChunkedBody(buf[:n])
		if err != nil {
			slog.Error("failed to write httpbin response body", "error", err, "path", path, "body", buf[n:])
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		slog.Error("failed to write httpbin response body", "error", err, "path", path)
		return
	}
	t := headers.NewHeaders()
	t.Set("X-Content-SHA256", fmt.Sprint(sha256.Sum256(body)))
	t.Set("X-Content-Length", fmt.Sprint(len(body)))
	err = w.WriteTrailers(t)
	if err != nil {
		slog.Error("failed to write trailers", "error", err, "path", path)
	}
}
