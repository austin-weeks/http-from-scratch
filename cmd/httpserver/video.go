package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/austin-weeks/http-from-scratch/internal/headers"
	"github.com/austin-weeks/http-from-scratch/internal/response"
)

func sendVideo(w *response.Writer) {
	v, err := os.Open("./assets/vim.mp4")
	if err != nil {
		slog.Error("failed to open video file", "error", err)
		_ = w.WriteStatusLine(404)
		h := headers.NewHeaders()
		h.Set("Connection", "close")
		_ = w.WriteHeaders(h)
		return
	}
	defer v.Close() // nolint

	s, err := v.Stat()
	if err != nil {
		panic(err)
	}

	_ = w.WriteStatusLine(response.StatusOK)

	h := headers.NewHeaders()
	h.Set("Connection", "close")
	h.Set("Content-Type", "video/mp4")
	h.Set("Content-Length", fmt.Sprint(s.Size()))

	_ = w.WriteHeaders(h)

	bodyLength := 0
	buf := make([]byte, 1024)
	for {
		n, readErr := v.Read(buf)
		bodyLength += n

		_, _ = w.WriteBody(buf[:n])

		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			} else {
				slog.Error("error reading video file", "error", readErr)
				return
			}
		}
	}
}
