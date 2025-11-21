// Package response provides methods for creating an HTTP response.
package response

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/austin-weeks/http-from-scratch/internal/headers"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GetDefaultHeaders(contentLength int) *headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprint(contentLength))
	h.Set("Content-Type", "text/plain")
	h.Set("Connection", "close")

	return h
}

type StatusCode int

const (
	StatusOK            StatusCode = 200
	StatusBadRequest    StatusCode = 400
	StatusInternalError StatusCode = 500
)

type Writer struct {
	conn io.Writer
}

func NewWriter(connection io.Writer) *Writer {
	return &Writer{
		conn: connection,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	reason := ""
	switch statusCode {
	case StatusOK:
		reason = "OK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusInternalError:
		reason = "Internal Server Error"
	}
	statusLine := fmt.Appendf(nil, "HTTP/1.1 %d %s\r\n", statusCode, reason)
	for len(statusLine) > 0 {
		n, err := w.conn.Write(statusLine)
		if err != nil {
			return err
		}
		statusLine = statusLine[n:]
	}
	return nil
}

func (w *Writer) WriteHeaders(headers *headers.Headers) error {
	var p []byte
	headers.ForEach(func(k, v string) {
		k = formatHeaderName(k)
		p = fmt.Appendf(p, "%s: %s\r\n", k, v)
	})
	p = append(p, []byte("\r\n")...)

	for len(p) > 0 {
		n, err := w.conn.Write(p)
		if err != nil {
			return err
		}
		p = p[n:]
	}
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	// don't mutate p
	written := 0
	for written < len(p) {
		n, err := w.conn.Write(p[written:])
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	lenHex := strconv.FormatInt(int64(len(p)), 16)
	body := fmt.Appendf(nil, "%s\r\n%s\r\n", lenHex, p)
	written := 0
	for written < len(body) {
		n, err := w.conn.Write(body[written:])
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	body := []byte("0\r\n\r\n")
	written := 0
	for written < len(body) {
		n, err := w.conn.Write(body[written:])
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
}

func (w *Writer) WriteTrailers(t *headers.Headers) error {
	return w.WriteHeaders(t)
}

func formatHeaderName(h string) string {
	c := cases.Title(language.English)
	parts := strings.Split(h, "-")
	for i, s := range parts {
		parts[i] = c.String(s)
	}
	return strings.Join(parts, "-")
}
