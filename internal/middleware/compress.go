package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

const (
	encoding        = "Content-Encoding"
	acceptEncoding  = "Accept-Encoding"
	contentType     = "Content-Type"
	contentLength   = "Content-Length"
	gzipEncoding    = "gzip"
	minCompressSize = 1024
)

var gzipPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz          *gzip.Writer
	writeHeader sync.Once
	statuscode  int
}

func (g *gzipResponseWriter) WriteHeader(code int) {
	g.writeHeader.Do(func() {
		g.statuscode = code
		g.Header().Del(contentLength)
		g.Header().Set(encoding, gzipEncoding)
		g.ResponseWriter.WriteHeader(code)
	})
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.statuscode == 0 {
		g.WriteHeader(http.StatusOK)
	}
	return g.gz.Write(b)
}

// NewCompress returns middleware that gzip-compresses responses when the
// client signals support via Accept-Encoding: gzip.
func NewCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get(acceptEncoding), gzipEncoding) {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(w)
		defer func() {
			_ = gz.Close()
			gzipPool.Put(gz)
		}()

		gw := &gzipResponseWriter{
			ResponseWriter: w,
			gz:             gz,
		}

		next.ServeHTTP(gw, r)
	})
}
