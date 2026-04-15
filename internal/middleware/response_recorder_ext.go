package middleware

import (
	"bytes"
	"net/http"
)

// bodyRecorder wraps ResponseRecorder to also capture the response body.
// It is used internally by the cache middleware.
type bodyRecorder struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

// NewResponseRecorder returns a bodyRecorder that delegates to the given ResponseWriter.
// It satisfies the interface used by both the logging and cache middleware.
func NewResponseRecorder(w http.ResponseWriter) *bodyRecorder {
	return &bodyRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (r *bodyRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *bodyRecorder) Write(b []byte) (int, error) {
	r.buf.Write(b) //nolint:errcheck
	return r.ResponseWriter.Write(b)
}

// Status returns the recorded HTTP status code.
func (r *bodyRecorder) Status() int {
	return r.status
}

// Body returns the recorded response body bytes.
func (r *bodyRecorder) Body() []byte {
	return r.buf.Bytes()
}
