package registry

import (
	"testing"
)

const sampleYAML = `cluster-01:
  - name: my-db
    namespace: default
    claimRef: postgresqls.2.2.2/my-db
    statusMessage: Ready
    lastCheckedAt: "2026-01-01T00:00:00Z"
  - name: my-cache
    namespace: default
    claimRef: redis.3.0.0/my-cache
    statusMessage: Pending
    lastCheckedAt: "2026-01-01T00:00:00Z"
cluster-02:
  - name: web-db
    namespace: apps
    claimRef: postgresqls.2.2.2/web-db
    statusMessage: Ready
    lastCheckedAt: "2026-01-01T00:00:00Z"
`

func TestParseRegistry(t *testing.T) {
	reg, err := ParseRegistry([]byte(sampleYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reg.Clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(reg.Clusters))
	}
	claims := reg.Clusters["cluster-01"]
	if len(claims) != 2 {
		t.Fatalf("expected 2 claims in cluster-01, got %d", len(claims))
	}
	if claims[0].ClaimRef != "postgresqls.2.2.2/my-db" {
		t.Errorf("unexpected claimRef: %s", claims[0].ClaimRef)
	}
}

func TestUpdateClaimStatus_Existing(t *testing.T) {
	reg, err := ParseRegistry([]byte(sampleYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	changed := UpdateClaimStatus(reg, "cluster-01", "postgresqls.2.2.2/my-db", "Degraded")
	if !changed {
		t.Fatal("expected UpdateClaimStatus to return true")
	}

	claim := reg.Clusters["cluster-01"][0]
	if claim.StatusMessage != "Degraded" {
		t.Errorf("expected StatusMessage 'Degraded', got %q", claim.StatusMessage)
	}
	if claim.LastCheckedAt == "2026-01-01T00:00:00Z" {
		t.Error("expected LastCheckedAt to be updated")
	}
}

func TestUpdateClaimStatus_NonExistent(t *testing.T) {
	reg, err := ParseRegistry([]byte(sampleYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	changed := UpdateClaimStatus(reg, "cluster-01", "nonexistent/claim", "Degraded")
	if changed {
		t.Fatal("expected UpdateClaimStatus to return false for non-existent claim")
	}

	changed = UpdateClaimStatus(reg, "no-such-cluster", "postgresqls.2.2.2/my-db", "Degraded")
	if changed {
		t.Fatal("expected UpdateClaimStatus to return false for non-existent cluster")
	}
}

func TestRoundTrip(t *testing.T) {
	reg1, err := ParseRegistry([]byte(sampleYAML))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	data, err := SerializeRegistry(reg1)
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	reg2, err := ParseRegistry(data)
	if err != nil {
		t.Fatalf("re-parse error: %v", err)
	}

	if len(reg2.Clusters) != len(reg1.Clusters) {
		t.Fatalf("cluster count mismatch: %d vs %d", len(reg1.Clusters), len(reg2.Clusters))
	}

	for cluster, claims1 := range reg1.Clusters {
		claims2, ok := reg2.Clusters[cluster]
		if !ok {
			t.Fatalf("cluster %q missing after round-trip", cluster)
		}
		if len(claims1) != len(claims2) {
			t.Fatalf("claim count mismatch for %q: %d vs %d", cluster, len(claims1), len(claims2))
		}
		for i := range claims1 {
			if claims1[i] != claims2[i] {
				t.Errorf("claim mismatch at %s[%d]: %+v vs %+v", cluster, i, claims1[i], claims2[i])
			}
		}
	}
}

func TestParseRegistry_Empty(t *testing.T) {
	reg, err := ParseRegistry([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reg.Clusters == nil {
		t.Fatal("expected non-nil Clusters map")
	}
	if len(reg.Clusters) != 0 {
		t.Fatalf("expected 0 clusters, got %d", len(reg.Clusters))
	}
}
