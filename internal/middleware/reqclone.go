package middleware

import (
	"bytes"
	"io"
	"net/http"
)

// NewRequestClone mirrors every incoming request body to a configured
// HTTP endpoint without affecting the primary response. Unlike shadow,
// this clones the request at the middleware layer before the body is
// consumed, making it safe to use alongside body-reading middleware.
func NewRequestClone(cloneURL string, client *http.Client) func(http.Handler) http.Handler {
	if client == nil {
		client = http.DefaultClient
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil && r.Body != http.NoBody {
				body, err := io.ReadAll(r.Body)
				if err == nil {
					r.Body = io.NopCloser(bytes.NewReader(body))
					go cloneRequest(client, r, cloneURL, body)
				}
			} else {
				go cloneRequest(client, r, cloneURL, nil)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func cloneRequest(client *http.Client, orig *http.Request, target string, body []byte) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(orig.Method, target, bodyReader)
	if err != nil {
		return
	}
	for k, vv := range orig.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("X-Cloned-From", orig.URL.Path)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}
