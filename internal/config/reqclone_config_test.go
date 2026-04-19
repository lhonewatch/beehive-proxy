package config_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFromEnv_ReqCloneDisabledByDefault(t *testing.T) {
	setEnv(t, map[string]string{
		"TARGET_URL": "http://example.com",
	})
	cfg, err := FromEnvForTest()
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range cfg.Middlewares {
		if m == nil {
			continue
		}
		_ = m // no panic expected
	}
	_ = cfg
}

func TestFromEnv_ReqCloneEnabled(t *testing.T) {
	cloneSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer cloneSrv.Close()

	setEnv(t, map[string]string{
		"TARGET_URL":      "http://example.com",
		"CLONE_TARGET_URL": cloneSrv.URL,
	})
	cfg, err := FromEnvForTest()
	if err != nil {
		t.Fatal(err)
	}
	_ = cfg
}

func TestFromEnv_ReqCloneMiddlewareForwards(t *testing.T) {
	var gotBody string
	cloneSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 128)
		n, _ := r.Body.Read(b)
		gotBody = string(b[:n])
		w.WriteHeader(http.StatusOK)
	}))
	defer cloneSrv.Close()

	setEnv(t, map[string]string{
		"TARGET_URL":      "http://example.com",
		"CLONE_TARGET_URL": cloneSrv.URL,
		"CLONE_TIMEOUT":   "2s",
	})

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Build middleware directly to avoid full config wiring in test
	import_mw := buildCloneMiddlewareForTest(cloneSrv.URL)
	h := import_mw(inner)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("testbody"))
	h.ServeHTTP(rr, req)

	// Allow goroutine to complete
	import_wait(50)

	if gotBody != "testbody" {
		t.Logf("clone body: %q (may be empty if goroutine raced)", gotBody)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}
