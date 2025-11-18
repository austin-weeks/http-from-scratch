// Package response provides methods for creating an HTTP response.
package response

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/austin-weeks/http-from-scratch/internal/headers"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type StatusCode int

const (
	StatusOK            StatusCode = 200
	StatusBadRequest    StatusCode = 400
	StatusInternalError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
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
	written := 0
	for written < len(statusLine) {
		n, err := w.Write(statusLine[written:])
		if err != nil {
			return err
		}
		written += n
	}
	return nil
}

func GetDefaultHeaders(contentLength int) headers.Headers {
	h := headers.Headers{}
	h["content-length"] = fmt.Sprint(contentLength)
	h["content-type"] = "text/plain"
	h["connection"] = "close"

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	if headers == nil {
		return errors.New("headers is nil")
	}

	var p []byte
	for k, v := range headers {
		k = formatHeaderName(k)
		p = fmt.Appendf(p, "%s: %s\r\n", k, v)
	}
	p = append(p, []byte("\r\n")...)

	for len(p) > 0 {
		n, err := w.Write(p)
		if err != nil {
			return err
		}
		p = p[n:]
	}
	return nil
}

func formatHeaderName(h string) string {
	c := cases.Title(language.English)
	parts := strings.Split(h, "-")
	for i, s := range parts {
		parts[i] = c.String(s)
	}
	return strings.Join(parts, "-")
}
