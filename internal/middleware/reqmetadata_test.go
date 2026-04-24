package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"testing"

	"github.com/sn3d/beehive-proxy/internal/middleware"
)

type capturedMetaHeaders struct {
	ServiceName    string
	ServiceVersion string
	RuntimeVersion string
	RuntimeCPUs    string
}

func captureMetaHandler(out *capturedMetaHeaders) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		out.ServiceName = r.Header.Get("X-Service-Name")
		out.ServiceVersion = r.Header.Get("X-Service-Version")
		out.RuntimeVersion = r.Header.Get("X-Runtime-Version")
		out.RuntimeCPUs = r.Header.Get("X-Runtime-CPUs")
		w.WriteHeader(http.StatusOK)
	})
}

func TestRequestMetadata_SetsDefaultServiceName(t *testing.T) {
	var got capturedMetaHeaders
	h := middleware.NewRequestMetadata(captureMetaHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got.ServiceName != "beehive-proxy" {
		t.Errorf("expected beehive-proxy, got %q", got.ServiceName)
	}
}

func TestRequestMetadata_SetsRuntimeVersion(t *testing.T) {
	var got capturedMetaHeaders
	h := middleware.NewRequestMetadata(captureMetaHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got.RuntimeVersion != runtime.Version() {
		t.Errorf("expected %q, got %q", runtime.Version(), got.RuntimeVersion)
	}
}

func TestRequestMetadata_SetsNumCPUs(t *testing.T) {
	var got capturedMetaHeaders
	h := middleware.NewRequestMetadata(captureMetaHandler(&got))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	want := strconv.Itoa(runtime.NumCPU())
	if got.RuntimeCPUs != want {
		t.Errorf("expected %q, got %q", want, got.RuntimeCPUs)
	}
}

func TestRequestMetadata_CustomServiceName(t *testing.T) {
	var got capturedMetaHeaders
	h := middleware.NewRequestMetadata(
		captureMetaHandler(&got),
		middleware.WithMetadataServiceName("my-service"),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got.ServiceName != "my-service" {
		t.Errorf("expected my-service, got %q", got.ServiceName)
	}
}

func TestRequestMetadata_CustomHeader(t *testing.T) {
	var captured string
	h := middleware.NewRequestMetadata(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			captured = r.Header.Get("X-Proxy-Name")
			w.WriteHeader(http.StatusOK)
		}),
		middleware.WithMetadataRequestHeader("X-Proxy-Name"),
		middleware.WithMetadataServiceName("edge"),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if captured != "edge" {
		t.Errorf("expected edge, got %q", captured)
	}
}

func TestRequestMetadata_CustomVersion(t *testing.T) {
	var got capturedMetaHeaders
	h := middleware.NewRequestMetadata(
		captureMetaHandler(&got),
		middleware.WithMetadataServiceVersion("v1.2.3"),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if got.ServiceVersion != "v1.2.3" {
		t.Errorf("expected v1.2.3, got %q", got.ServiceVersion)
	}
}
