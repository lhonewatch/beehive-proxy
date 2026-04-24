package middleware

import (
	"net/http"
	"sync"
	"time"
)

// burstWindow tracks request counts within a rolling window.
type burstWindow struct {
	mu       sync.Mutex
	counts   []int64
	buckets  int
	bucket   time.Duration
	current  int
	last     time.Time
}

func newBurstWindow(window time.Duration, buckets int) *burstWindow {
	return &burstWindow{
		counts:  make([]int64, buckets),
		buckets: buckets,
		bucket:  window / time.Duration(buckets),
		last:    time.Now(),
	}
}

func (bw *burstWindow) add() int64 {
	bw.mu.Lock()
	defer bw.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(bw.last)
	shift := int(elapsed / bw.bucket)
	if shift > 0 {
		for i := 0; i < shift && i < bw.buckets; i++ {
			bw.current = (bw.current + 1) % bw.buckets
			bw.counts[bw.current] = 0
		}
		bw.last = now
	}
	bw.counts[bw.current]++
	var total int64
	for _, c := range bw.counts {
		total += c
	}
	return total
}

// NewRequestBurst returns a middleware that limits the total number of requests
// within a rolling time window. Requests exceeding the burst limit receive 429.
func NewRequestBurst(limit int64, window time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	windows := map[string]*burstWindow{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)
			mu.Lock()
			bw, ok := windows[ip]
			if !ok {
				bw = newBurstWindow(window, 10)
				windows[ip] = bw
			}
			mu.Unlock()

			count := bw.add()
			if limit > 0 && count > limit {
				http.Error(w, "burst limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
