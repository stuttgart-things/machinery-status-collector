package collector

import (
	"sync"
	"time"
)

// StatusEntry represents a status update received from a cluster agent.
type StatusEntry struct {
	Cluster       string
	ClaimRef      string
	StatusMessage string
	ReceivedAt    time.Time
}

// StatusStore is a thread-safe in-memory store for collecting status updates.
// It tracks dirty state to signal when a reconciliation PR is needed.
type StatusStore struct {
	sync.RWMutex
	entries map[string]StatusEntry
	dirty   bool
}

// NewStatusStore creates an empty StatusStore.
func NewStatusStore() *StatusStore {
	return &StatusStore{
		entries: make(map[string]StatusEntry),
	}
}

func storeKey(cluster, claimRef string) string {
	return cluster + "/" + claimRef
}

// Put inserts or updates a status entry and marks the store as dirty.
func (s *StatusStore) Put(cluster, claimRef, status string) {
	s.Lock()
	defer s.Unlock()
	s.entries[storeKey(cluster, claimRef)] = StatusEntry{
		Cluster:       cluster,
		ClaimRef:      claimRef,
		StatusMessage: status,
		ReceivedAt:    time.Now().UTC(),
	}
	s.dirty = true
}

// Get retrieves a status entry by cluster and claimRef.
func (s *StatusStore) Get(cluster, claimRef string) (StatusEntry, bool) {
	s.RLock()
	defer s.RUnlock()
	e, ok := s.entries[storeKey(cluster, claimRef)]
	return e, ok
}

// GetAll returns a snapshot copy of all entries in the store.
func (s *StatusStore) GetAll() []StatusEntry {
	s.RLock()
	defer s.RUnlock()
	result := make([]StatusEntry, 0, len(s.entries))
	for _, e := range s.entries {
		result = append(result, e)
	}
	return result
}

// IsDirty reports whether the store has been modified since the last flush.
func (s *StatusStore) IsDirty() bool {
	s.RLock()
	defer s.RUnlock()
	return s.dirty
}

// MarkFlushed resets the dirty flag.
func (s *StatusStore) MarkFlushed() {
	s.Lock()
	defer s.Unlock()
	s.dirty = false
}
