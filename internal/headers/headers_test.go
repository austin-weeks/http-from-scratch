package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParsing(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("Content-Type:   application/json   \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "application/json", headers.Get("Content-Type"))
	assert.Equal(t, n, 37)
	assert.False(t, done)

	// Test: Valid 2 headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nContent-Type: application/json\r\n")
	n, _, err = headers.Parse(data)
	require.NoError(t, err)
	_, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "application/json", headers.Get("content-type"))

	// Test: Valid done
	headers = NewHeaders()
	_, done, err = headers.Parse([]byte("\r\n"))
	require.NoError(t, err)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid characters
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	_, _, err = headers.Parse(data)
	require.Error(t, err)

	// Test: Duplicate headers
	headers = NewHeaders()
	headers.Set("Accept", "application/json")
	_, _, err = headers.Parse([]byte("Accept: xml\r\n"))
	require.NoError(t, err)
	assert.Equal(t, "application/json, xml", headers.Get("Accept"))
}
