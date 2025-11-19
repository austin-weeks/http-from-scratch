// Package headers provides a Headers type representing HTTP headers.
package headers

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	crlf           = []byte("\r\n")
	fieldNameRegex = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+-./^_`|~]+$")
)

type Headers struct {
	hMap map[string]string
}

func NewHeaders() *Headers {
	return &Headers{
		hMap: make(map[string]string),
	}
}

func (h *Headers) Get(key string) string {
	return h.hMap[strings.ToLower(key)]
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)
	if v, ok := h.hMap[name]; ok {
		value = fmt.Sprintf("%s, %s", v, value)
	}

	h.hMap[name] = value
}

func (h *Headers) OverwriteSet(name, value string) {
	name = strings.ToLower(name)
	h.hMap[name] = value
}

func (h *Headers) ForEach(fn func(k, v string)) {
	for k, v := range h.hMap {
		fn(k, v)
	}
}

func (h *Headers) Parse(data []byte) (n int, done bool, err error) {
	i := bytes.Index(data, crlf)
	if i == -1 {
		return 0, false, nil
	}
	// End of headers
	if i == 0 {
		return len(crlf), true, nil
	}

	read := i + len(crlf)
	header := data[:i]
	header = bytes.TrimSpace(header)

	i = bytes.Index(header, []byte(":"))
	if i == -1 {
		return 0, false, errors.New("no colon found in header line")
	}
	name, value := header[:i], header[i+1:]
	if !fieldNameRegex.Match(name) {
		return 0, false, errors.New("field name contains invalid characters")
	}
	nameStr := string(bytes.ToLower(name))
	valueStr := string(bytes.TrimSpace(value))

	h.Set(nameStr, valueStr)

	return read, false, nil
}
