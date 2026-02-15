package informer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestExtractClaimStatus_ReadyCondition(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "example.org/v1",
			"kind":       "PostgreSQL",
			"metadata": map[string]interface{}{
				"name":      "my-db",
				"namespace": "default",
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":    "Synced",
						"status":  "True",
						"message": "Resource is synced",
					},
					map[string]interface{}{
						"type":    "Ready",
						"status":  "True",
						"message": "Resource is available",
					},
				},
			},
		},
	}

	msg, err := ExtractClaimStatus(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "Resource is available" {
		t.Errorf("expected 'Resource is available', got %q", msg)
	}
}

func TestExtractClaimStatus_ReadyNoMessage(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "False",
					},
				},
			},
		},
	}

	msg, err := ExtractClaimStatus(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "Ready=False" {
		t.Errorf("expected 'Ready=False', got %q", msg)
	}
}

func TestExtractClaimStatus_NoConditions(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{},
		},
	}

	msg, err := ExtractClaimStatus(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "no status conditions available" {
		t.Errorf("expected 'no status conditions available', got %q", msg)
	}
}

func TestExtractClaimStatus_NoStatus(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{},
	}

	msg, err := ExtractClaimStatus(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "no status conditions available" {
		t.Errorf("expected 'no status conditions available', got %q", msg)
	}
}

func TestExtractClaimStatus_NoReadyCondition(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":    "Synced",
						"status":  "True",
						"message": "Resource is synced",
					},
				},
			},
		},
	}

	msg, err := ExtractClaimStatus(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "no Ready condition found" {
		t.Errorf("expected 'no Ready condition found', got %q", msg)
	}
}

func TestSendStatus_RequestFormat(t *testing.T) {
	var received statusPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}
		if r.URL.Path != "/api/v1/status" {
			t.Errorf("expected path /api/v1/status, got %q", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	gvr := schema.GroupVersionResource{Group: "example.org", Version: "v1", Resource: "postgresqls"}
	w := NewClaimWatcher(nil, ts.URL, "cluster-01", gvr, "default")

	claim := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "example.org/v1",
			"kind":       "PostgreSQL",
			"metadata": map[string]interface{}{
				"name":      "my-db",
				"namespace": "default",
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":    "Ready",
						"status":  "True",
						"message": "Resource is available",
					},
				},
			},
		},
	}

	if err := w.sendStatus(claim); err != nil {
		t.Fatalf("sendStatus failed: %v", err)
	}

	if received.Cluster != "cluster-01" {
		t.Errorf("expected cluster 'cluster-01', got %q", received.Cluster)
	}
	if received.ClaimRef != "default/my-db" {
		t.Errorf("expected claimRef 'default/my-db', got %q", received.ClaimRef)
	}
	if received.StatusMessage != "Resource is available" {
		t.Errorf("expected statusMessage 'Resource is available', got %q", received.StatusMessage)
	}
}

func TestSendStatus_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	gvr := schema.GroupVersionResource{Group: "example.org", Version: "v1", Resource: "postgresqls"}
	w := NewClaimWatcher(nil, ts.URL, "cluster-01", gvr, "default")

	claim := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "my-db",
				"namespace": "default",
			},
			"status": map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":    "Ready",
						"status":  "True",
						"message": "Ready",
					},
				},
			},
		},
	}

	err := w.sendStatus(claim)
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}
