package traefik_request_id

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var uuidV4Re = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewUUID_Format(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := newUUID()
		if !uuidV4Re.MatchString(id) {
			t.Fatalf("invalid UUID v4: %s", id)
		}
	}
}

func TestNewUUID_Uniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 1000)

	for i := 0; i < 1000; i++ {
		id := newUUID()
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate UUID: %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestServeHTTP_GeneratesID(t *testing.T) {
	var capturedHeader string

	next := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		capturedHeader = req.Header.Get("X-Request-ID")
	})

	handler, err := New(context.Background(), next, CreateConfig(), "test")
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if !uuidV4Re.MatchString(capturedHeader) {
		t.Fatalf("expected UUID v4 on request, got: %q", capturedHeader)
	}

	resp := rec.Header().Get("X-Request-ID")
	if resp != capturedHeader {
		t.Fatalf("response header %q != request header %q", resp, capturedHeader)
	}
}

func TestServeHTTP_OverridesClientID(t *testing.T) {
	clientID := "client-provided-id"
	var capturedHeader string

	next := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		capturedHeader = req.Header.Get("X-Request-ID")
	})

	handler, err := New(context.Background(), next, CreateConfig(), "test")
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", clientID)

	handler.ServeHTTP(rec, req)

	if capturedHeader == clientID {
		t.Fatal("client-provided header should have been overridden")
	}
	if !uuidV4Re.MatchString(capturedHeader) {
		t.Fatalf("expected UUID v4, got: %q", capturedHeader)
	}
	if rec.Header().Get("X-Request-ID") != capturedHeader {
		t.Fatalf("response header should match generated ID")
	}
}

func TestServeHTTP_OverridesBackendID(t *testing.T) {
	var capturedHeader string

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		capturedHeader = req.Header.Get("X-Request-ID")
		rw.Header().Set("X-Request-ID", "backend-id")
	})

	handler, err := New(context.Background(), next, CreateConfig(), "test")
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	resp := rec.Header().Get("X-Request-ID")
	if resp != capturedHeader {
		t.Fatalf("response header %q should match generated ID %q, not backend value", resp, capturedHeader)
	}
}

func TestServeHTTP_UniquePerRequest(t *testing.T) {
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

	handler, err := New(context.Background(), next, CreateConfig(), "test")
	if err != nil {
		t.Fatal(err)
	}

	seen := make(map[string]struct{}, 100)

	for i := 0; i < 100; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		handler.ServeHTTP(rec, req)

		id := rec.Header().Get("X-Request-ID")
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate request ID across requests: %s", id)
		}
		seen[id] = struct{}{}
	}
}

func TestServeHTTP_CustomHeaderName(t *testing.T) {
	cfg := &Config{HeaderName: "X-Correlation-ID"}
	var capturedHeader string

	next := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		capturedHeader = req.Header.Get("X-Correlation-ID")
	})

	handler, err := New(context.Background(), next, cfg, "test")
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.ServeHTTP(rec, req)

	if !uuidV4Re.MatchString(capturedHeader) {
		t.Fatalf("expected UUID v4 on custom header, got: %q", capturedHeader)
	}

	if rec.Header().Get("X-Correlation-ID") != capturedHeader {
		t.Fatalf("response should use custom header name")
	}
}

func TestNew_DefaultsEmptyHeaderName(t *testing.T) {
	cfg := &Config{HeaderName: ""}

	handler, err := New(context.Background(), http.NotFoundHandler(), cfg, "test")
	if err != nil {
		t.Fatal(err)
	}

	rid := handler.(*requestID)
	if rid.headerName != defaultHeaderName {
		t.Fatalf("expected %q, got %q", defaultHeaderName, rid.headerName)
	}
}

func TestHexEncode(t *testing.T) {
	src := []byte{0x00, 0xff, 0x0a, 0xbc}
	dst := make([]byte, 8)
	hexEncode(dst, src)

	if string(dst) != "00ff0abc" {
		t.Fatalf("expected 00ff0abc, got %s", string(dst))
	}
}

func BenchmarkNewUUID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		newUUID()
	}
}

func BenchmarkServeHTTP(b *testing.B) {
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	handler, _ := New(context.Background(), next, CreateConfig(), "bench")
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
