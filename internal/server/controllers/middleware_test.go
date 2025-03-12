package controllers

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testResponseWriter struct {
	*httptest.ResponseRecorder
}

func (w *testResponseWriter) WriteHeaderNow()          {}
func (w *testResponseWriter) Size() int                { return w.Body.Len() }
func (w *testResponseWriter) Written() bool            { return w.Code != 0 }
func (w *testResponseWriter) Status() int              { return w.Code }
func (w *testResponseWriter) CloseNotify() <-chan bool { return make(chan bool) }
func (w *testResponseWriter) Flush()                   {}
func (w *testResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, http.ErrNotSupported
}
func (w *testResponseWriter) Pusher() http.Pusher {
	return nil
}

func TestGzipResponseWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	rec := httptest.NewRecorder()
	wrappedRec := &testResponseWriter{rec}

	gzw := &gzipResponseWriter{ResponseWriter: wrappedRec, writer: gz, buffer: &buf}

	data := []byte("Test data!")

	n, err := gzw.Write(data)

	require.NoError(t, err, "Write() returned an error")
	assert.Equal(t, len(data), n, "Write() returned an incorrect number of bytes")

	gz.Close()

	assert.NotEqual(t, data, buf.Bytes(), "Data was not compressed")
}

func TestGzipResponseWriter_Close(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	rec := httptest.NewRecorder()
	wrappedRec := &testResponseWriter{rec}

	gzw := &gzipResponseWriter{ResponseWriter: wrappedRec, writer: gz, buffer: &buf}

	gzw.Close()

	_, err := gz.Write([]byte("test"))
	assert.Error(t, err, "Write() did not return an error after closing the writer")
}
