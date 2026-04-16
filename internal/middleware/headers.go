package middleware

import "net/http"

// HeadersOptions configures static request/response header injection.
type HeadersOptions struct {
	// RequestHeaders are added to every proxied request.
	RequestHeaders map[string]string
	// ResponseHeaders are added to every response.
	ResponseHeaders map[string]string
	// RemoveRequestHeaders lists headers to strip from the incoming request.
	RemoveRequestHeaders []string
	// RemoveResponseHeaders lists headers to strip from the response.
	RemoveResponseHeaders []string
}

// NewHeaders returns middleware that injects and removes HTTP headers.
func NewHeaders(opts HeadersOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Mutate a clone so we don't modify the original.
			rc := r.Clone(r.Context())

			for _, h := range opts.RemoveRequestHeaders {
				rc.Header.Del(h)
			}
			for k, v := range opts.RequestHeaders {
				rc.Header.Set(k, v)
			}

			rw := &headerResponseWriter{ResponseWriter: w, opts: opts}
			next.ServeHTTP(rw, rc)
		})
	}
}

type headerResponseWriter struct {
	http.ResponseWriter
	opts    HeadersOptions
	wroteHeader bool
}

func (h *headerResponseWriter) WriteHeader(code int) {
	if !h.wroteHeader {
		h.applyResponseHeaders()
		h.wroteHeader = true
	}
	h.ResponseWriter.WriteHeader(code)
}

func (h *headerResponseWriter) Write(b []byte) (int, error) {
	if !h.wroteHeader {
		h.applyResponseHeaders()
		h.wroteHeader = true
	}
	return h.ResponseWriter.Write(b)
}

func (h *headerResponseWriter) applyResponseHeaders() {
	hdr := h.ResponseWriter.Header()
	for _, k := range h.opts.RemoveResponseHeaders {
		hdr.Del(k)
	}
	for k, v := range h.opts.ResponseHeaders {
		hdr.Set(k, v)
	}
}
