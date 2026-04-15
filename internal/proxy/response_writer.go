package proxy

import "net/http"

// responseWriter wraps http.ResponseWriter to capture the status code
// written by the upstream handler.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code before delegating to the
// underlying ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

// Write ensures that if WriteHeader has not been called yet, the
// default 200 OK status is recorded.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}
