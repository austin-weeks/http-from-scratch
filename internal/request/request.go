// Package request provides methods for parsing HTTP requests.
package request

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/austin-weeks/http-from-scratch/internal/headers"
)

type parseState int

const bufferSize = 8

var crlf = []byte("\r\n")

const (
	initialized parseState = iota
	parsingHeaders
	parsingBody
	done
)

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	Body        []byte
	state       parseState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	r := &Request{
		state:   initialized,
		Headers: headers.NewHeaders(),
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
				if r.state == parsingBody {
					return nil, errors.New("body shorter than reported Content-Length")
				} else {
					break
				}
			} else {
				return nil, readErr
			}
		}
	}

	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
loop:
	for {
		curr := data[read:]
		if len(curr) == 0 {
			break loop
		}
		switch r.state {
		case initialized:
			rl, n, err := parseRequestLine(curr)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break loop
			}
			r.RequestLine = *rl
			r.state = parsingHeaders
			read += n

		case parsingHeaders:
			n, doneParsing, err := r.Headers.Parse(curr)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break loop
			}
			read += n
			if doneParsing {
				if r.hasBody() {
					r.state = parsingBody
				} else {
					r.state = done
				}
			}

		case parsingBody:
			cl := r.getContentLength()
			bodyDataLen := len(curr)
			read += bodyDataLen
			r.Body = append(r.Body, curr...)
			bodyLen := len(r.Body)
			if bodyLen > cl {
				return 0, errors.New("body longer than reported Content-Length")
			} else if bodyLen == cl {
				r.state = done
			}

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

func (r Request) hasBody() bool {
	return r.getContentLength() > 0
}

func (r Request) getContentLength() int {
	clheader := r.Headers.Get("Content-Length")
	if clheader == "" {
		return 0
	}
	cl, err := strconv.Atoi(clheader)
	if err != nil {
		return 0
	}
	return cl
}
