package server

import (
	"fmt"
	"io"

	"github.com/austin-weeks/http-from-scratch/internal/request"
	"github.com/austin-weeks/http-from-scratch/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (e HandlerError) writeTo(w io.Writer) error {
	p := fmt.Appendf(nil, "Error %d: %s", e.StatusCode, e.Message)
	for len(p) > 0 {
		n, err := w.Write(p)
		if err != nil {
			return err
		}
		p = p[n:]
	}
	return nil
}
