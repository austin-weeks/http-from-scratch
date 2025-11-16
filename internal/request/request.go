// Package request provides methods for parsing HTTP requests.
package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

type parseState int

const bufferSize = 8

var crlf = []byte("\r\n")

const (
	initialized parseState = iota
	done
)

type Request struct {
	RequestLine RequestLine
	state       parseState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := &Request{
		state: initialized,
	}
	buf := make([]byte, bufferSize)
	bufLen := 0

	for r.state != done {
		if bufLen == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		read, readErr := reader.Read(buf[bufLen:])
		bufLen += read

		parsed, err := r.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[parsed:bufLen])
		bufLen -= parsed

		// Handle non-EOF errors
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				r.state = done
				break
			} else {
				return nil, err
			}
		}
	}

	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
loop:
	for {
		switch r.state {
		case initialized:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break loop
			}
			r.RequestLine = *rl
			r.state = done
			read += n

		case done:
			break loop

		default:
			return 0, errors.New("unknown parse state")
		}
	}
	return read, nil
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	i := bytes.Index(b, crlf)
	if i == -1 {
		return nil, 0, nil
	}

	header := b[:i]
	read := i + len(crlf)

	parts := strings.Split(string(header), " ")
	if len(parts) != 3 {
		return nil, 0, errors.New("request line must have exactly 3 parts")
	}

	m, t, v := parts[0], parts[1], parts[2]
	if v != "HTTP/1.1" {
		return nil, 0, errors.New("HTTP version must be 'HTTP/1.1'")
	}
	v = strings.TrimPrefix(v, "HTTP/")

	if m != strings.ToUpper(m) {
		return nil, 0, errors.New("HTTP method must be uppercase")
	}

	return &RequestLine{
		HttpVersion:   v,
		RequestTarget: t,
		Method:        m,
	}, read, nil
}
