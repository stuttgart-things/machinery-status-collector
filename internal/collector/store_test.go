package collector

import (
	"sync"
	"testing"
)

func TestPutAndGet(t *testing.T) {
	s := NewStatusStore()

	s.Put("cluster-01", "postgresqls.2.2.2/my-db", "Ready")

	e, ok := s.Get("cluster-01", "postgresqls.2.2.2/my-db")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Cluster != "cluster-01" {
		t.Errorf("expected Cluster 'cluster-01', got %q", e.Cluster)
	}
	if e.ClaimRef != "postgresqls.2.2.2/my-db" {
		t.Errorf("expected ClaimRef 'postgresqls.2.2.2/my-db', got %q", e.ClaimRef)
	}
	if e.StatusMessage != "Ready" {
		t.Errorf("expected StatusMessage 'Ready', got %q", e.StatusMessage)
	}
	if e.ReceivedAt.IsZero() {
		t.Error("expected ReceivedAt to be set")
	}

	_, ok = s.Get("cluster-01", "nonexistent")
	if ok {
		t.Error("expected entry to not exist")
	}
}

func TestDirtyFlag(t *testing.T) {
	s := NewStatusStore()

	if s.IsDirty() {
		t.Fatal("new store should not be dirty")
	}

	s.Put("cluster-01", "postgresqls.2.2.2/my-db", "Ready")
	if !s.IsDirty() {
		t.Fatal("store should be dirty after Put")
	}

	s.MarkFlushed()
	if s.IsDirty() {
		t.Fatal("store should not be dirty after MarkFlushed")
	}

	s.Put("cluster-01", "postgresqls.2.2.2/my-db", "Degraded")
	if !s.IsDirty() {
		t.Fatal("store should be dirty after second Put")
	}
}

func TestConcurrentPut(t *testing.T) {
	s := NewStatusStore()
	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Put("cluster-01", "claim/ref", "status")
			s.Get("cluster-01", "claim/ref")
			s.IsDirty()
			s.GetAll()
		}()
	}

	wg.Wait()

	e, ok := s.Get("cluster-01", "claim/ref")
	if !ok {
		t.Fatal("expected entry to exist after concurrent writes")
	}
	if e.StatusMessage != "status" {
		t.Errorf("unexpected StatusMessage: %q", e.StatusMessage)
	}
}

func TestGetAllReturnsSnapshot(t *testing.T) {
	s := NewStatusStore()
	s.Put("cluster-01", "claim/a", "Ready")
	s.Put("cluster-02", "claim/b", "Pending")

	snapshot := s.GetAll()
	if len(snapshot) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snapshot))
	}

	// Mutating the returned slice should not affect the store.
	snapshot[0].StatusMessage = "MODIFIED"

	e, _ := s.Get(snapshot[0].Cluster, snapshot[0].ClaimRef)
	if e.StatusMessage == "MODIFIED" {
		t.Error("GetAll must return a copy; store was mutated via returned slice")
	}
}
