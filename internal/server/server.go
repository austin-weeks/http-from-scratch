// Package server provides an HTTP server.
package server

import (
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

	w := response.NewWriter(conn)
	s.handler(w, r)
}
