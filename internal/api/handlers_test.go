package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stuttgart-things/machinery-status-collector/internal/collector"
)

func newTestServer() *Server {
	return NewServer(collector.NewStatusStore(), "v0.1.0-test", "abc1234")
}

func TestPostStatus_Valid(t *testing.T) {
	srv := newTestServer()

	body := `{"cluster":"cluster-a","claimRef":"my/claim","statusMessage":"ready"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	entry, ok := srv.store.Get("cluster-a", "my/claim")
	if !ok {
		t.Fatal("expected entry in store")
	}
	if entry.StatusMessage != "ready" {
		t.Fatalf("expected status 'ready', got %q", entry.StatusMessage)
	}
}

func TestPostStatus_InvalidJSON(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/status", strings.NewReader("{invalid"))
	rec := httptest.NewRecorder()

	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPostStatus_MissingFields(t *testing.T) {
	srv := newTestServer()

	cases := []struct {
		name string
		body string
	}{
		{"missing cluster", `{"claimRef":"c","statusMessage":"ok"}`},
		{"missing claimRef", `{"cluster":"a","statusMessage":"ok"}`},
		{"missing statusMessage", `{"cluster":"a","claimRef":"c"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/status", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			srv.Handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", rec.Code)
			}
		})
	}
}

func TestGetStatus(t *testing.T) {
	srv := newTestServer()
	srv.store.Put("cluster-a", "claim1", "ready")
	srv.store.Put("cluster-b", "claim2", "pending")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp []statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp))
	}
}

func TestGetStatusByCluster(t *testing.T) {
	srv := newTestServer()
	srv.store.Put("cluster-a", "claim1", "ready")
	srv.store.Put("cluster-a", "claim2", "pending")
	srv.store.Put("cluster-b", "claim3", "failed")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status/cluster-a", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp []statusResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 entries for cluster-a, got %d", len(resp))
	}
	for _, r := range resp {
		if r.Cluster != "cluster-a" {
			t.Fatalf("expected cluster 'cluster-a', got %q", r.Cluster)
		}
	}
}

func TestHealthz(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("expected status 'ok', got %q", resp["status"])
	}
}

func TestVersion(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["version"] != "v0.1.0-test" {
		t.Fatalf("expected version 'v0.1.0-test', got %q", resp["version"])
	}
	if resp["commit"] != "abc1234" {
		t.Fatalf("expected commit 'abc1234', got %q", resp["commit"])
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	srv := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	// Build a handler that panics, wrapped with recovery middleware.
	panicker := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
	handler := wrapMiddleware(panicker, recoveryMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}
