package middleware

import (
	"context"
	"net/http"
	"time"
)

// NewTimeout returns a middleware that cancels the request context after the
// given duration. If the handler does not respond in time, a 504 Gateway
// Timeout is written and the context is cancelled so upstream transports can
// abort in-flight requests.
//
// Note: if the handler has already started writing to the ResponseRecorder
// before the timeout fires, the partial response is discarded and the 504 is
// sent instead.
func NewTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan struct{})
			rw := NewResponseRecorder(w)

			go func() {
				defer close(done)
				next.ServeHTTP(rw, r)
			}()

			select {
			case <-done:
				// Handler finished in time — flush the recorded response.
				w.WriteHeader(rw.Status())
				_, _ = w.Write(rw.Body())
			case <-ctx.Done():
				// Deadline exceeded; surface a clear timeout error to the caller.
				// We must wait for the goroutine to finish before returning so
				// that we don't leak it beyond the handler's lifetime.
				http.Error(w, "gateway timeout", http.StatusGatewayTimeout)
				<-done
			}
		})
	}
}
