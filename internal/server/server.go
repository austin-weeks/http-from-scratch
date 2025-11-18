// Package server provides an HTTP server.
package server

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync/atomic"

	"github.com/austin-weeks/http-from-scratch/internal/request"
	"github.com/austin-weeks/http-from-scratch/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

func Serve(port uint16, handler Handler) (*Server, error) {
	if handler == nil {
		return nil, errors.New("handler function cannot be nil")
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		listener: l,
		handler:  handler,
	}
	go s.listen()
	return s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	err := s.listener.Close()
	return err
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			slog.Error("error accepting connection", "error", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close() // nolint

	r, err := request.RequestFromReader(conn)
	if err != nil {
		slog.Error("failed to read request", "connection", conn, "error", err)
		return
	}

	status := response.StatusOK
	var body bytes.Buffer
	handlerErr := s.handler(&body, r)
	if handlerErr != nil {
		status = handlerErr.StatusCode
		body = *bytes.NewBuffer([]byte(handlerErr.Message))
	}

	headers := response.GetDefaultHeaders(body.Len())
	err = response.WriteStatusLine(conn, status)
	if err != nil {
		slog.Error("failed to write response status line", "connection", conn, "request", r, "error", err)
		return
	}

	err = response.WriteHeaders(conn, headers)
	if err != nil {
		slog.Error("failed to write response headers", "connection", conn, "request", r, "error", err)
		return
	}

	_, err = body.WriteTo(conn)
	if err != nil {
		slog.Error("failed to write response body", "connection", conn, "request", r, "error", err)
	}
}
