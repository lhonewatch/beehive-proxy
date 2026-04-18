package middleware

import (
	"net/http"
	"sync/atomic"
)

// DuplexOptions configures request mirroring behaviour.
type DuplexOptions struct {
	// MirrorURL is the base URL to which a copy of each request is sent.
	MirrorURL string
	// Client is the HTTP client used for mirror requests. Defaults to http.DefaultClient.
	Client *http.Client
}

type duplexMiddleware struct {
	opts    DuplexOptions
	client  *http.Client
	Dropped uint64 // atomic counter of mirror errors
}

// NewDuplex returns middleware that forwards a shadow copy of every request to
// opts.MirrorURL without affecting the primary response.
func NewDuplex(opts DuplexOptions) func(http.Handler) http.Handler {
	client := opts.Client
	if client == nil {
		client = http.DefaultClient
	}
	dm := &duplexMiddleware{opts: opts, client: client}
	return dm.handler
}

func (d *duplexMiddleware) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		go d.mirror(r)
		next.ServeHTTP(w, r)
	})
}

func (d *duplexMiddleware) mirror(r *http.Request) {
	url := d.opts.MirrorURL + r.RequestURI
	req, err := http.NewRequest(r.Method, url, nil)
	if err != nil {
		atomic.AddUint64(&d.Dropped, 1)
		return
	}
	for k, v := range r.Header {
		req.Header[k] = v
	}
	req.Header.Set("X-Mirror", "1")
	resp, err := d.client.Do(req)
	if err != nil {
		atomic.AddUint64(&d.Dropped, 1)
		return
	}
	resp.Body.Close()
}
