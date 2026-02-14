package collector

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type mockGitClient struct {
	fetchFileContent []byte
	fetchFileSHA     string
	fetchFileErr     error

	getRefSHA string
	getRefErr error

	createBranchErr error

	updateFileErr error

	createPRNumber int
	createPRErr    error

	listOpenPRsNumbers []int
	listOpenPRsErr     error

	// Track calls for assertions.
	fetchFileCalled  bool
	getRefCalled     bool
	createBranchName string
	updateFileCalled bool
	createPRCalled   bool
	listOpenPRsCalled bool
}

func (m *mockGitClient) FetchFile(path, ref string) ([]byte, string, error) {
	m.fetchFileCalled = true
	return m.fetchFileContent, m.fetchFileSHA, m.fetchFileErr
}

func (m *mockGitClient) GetRef(branch string) (string, error) {
	m.getRefCalled = true
	return m.getRefSHA, m.getRefErr
}

func (m *mockGitClient) CreateBranch(baseSHA, branchName string) error {
	m.createBranchName = branchName
	return m.createBranchErr
}

func (m *mockGitClient) UpdateFile(path, branchName, message string, content []byte, sha string) error {
	m.updateFileCalled = true
	return m.updateFileErr
}

func (m *mockGitClient) CreatePR(title, body, head, base string) (int, error) {
	m.createPRCalled = true
	return m.createPRNumber, m.createPRErr
}

func (m *mockGitClient) ListOpenPRs(head string) ([]int, error) {
	m.listOpenPRsCalled = true
	return m.listOpenPRsNumbers, m.listOpenPRsErr
}

const testRegistryYAML = `cluster-a:
  - name: my-claim
    namespace: default
    claimRef: my-claim-ref
    statusMessage: pending
    lastCheckedAt: ""
`

func TestReconcileOnce_DirtyStore(t *testing.T) {
	store := NewStatusStore()
	store.Put("cluster-a", "my-claim-ref", "ready")

	mock := &mockGitClient{
		fetchFileContent:   []byte(testRegistryYAML),
		fetchFileSHA:       "filesha123",
		getRefSHA:          "commitsha456",
		listOpenPRsNumbers: []int{},
		createPRNumber:     7,
	}

	rec := NewReconciler(store, mock, time.Minute, "registry.yaml", "main")

	if err := rec.reconcileOnce(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.fetchFileCalled {
		t.Fatal("expected FetchFile to be called")
	}
	if !mock.getRefCalled {
		t.Fatal("expected GetRef to be called")
	}
	if mock.createBranchName == "" {
		t.Fatal("expected CreateBranch to be called")
	}
	if !mock.updateFileCalled {
		t.Fatal("expected UpdateFile to be called")
	}
	if !mock.createPRCalled {
		t.Fatal("expected CreatePR to be called")
	}
	if store.IsDirty() {
		t.Fatal("expected store to be flushed after successful reconcile")
	}
}

func TestReconcileOnce_CleanStore(t *testing.T) {
	store := NewStatusStore()

	mock := &mockGitClient{}

	rec := NewReconciler(store, mock, time.Minute, "registry.yaml", "main")

	if err := rec.reconcileOnce(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.fetchFileCalled {
		t.Fatal("expected no git calls when store is clean")
	}
}

func TestReconcileOnce_DuplicatePR(t *testing.T) {
	store := NewStatusStore()
	store.Put("cluster-a", "my-claim-ref", "ready")

	mock := &mockGitClient{
		fetchFileContent:   []byte(testRegistryYAML),
		fetchFileSHA:       "filesha123",
		listOpenPRsNumbers: []int{42},
	}

	rec := NewReconciler(store, mock, time.Minute, "registry.yaml", "main")

	if err := rec.reconcileOnce(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.listOpenPRsCalled {
		t.Fatal("expected ListOpenPRs to be called")
	}
	if mock.getRefCalled {
		t.Fatal("expected GetRef NOT to be called when duplicate PR exists")
	}
	if mock.createPRCalled {
		t.Fatal("expected CreatePR NOT to be called when duplicate PR exists")
	}
	if !store.IsDirty() {
		t.Fatal("expected store to remain dirty when PR creation is skipped")
	}
}

func TestReconcileOnce_FetchFileError(t *testing.T) {
	store := NewStatusStore()
	store.Put("cluster-a", "my-claim-ref", "ready")

	mock := &mockGitClient{
		fetchFileErr: fmt.Errorf("network error"),
	}

	rec := NewReconciler(store, mock, time.Minute, "registry.yaml", "main")

	err := rec.reconcileOnce(context.Background())
	if err == nil {
		t.Fatal("expected error when FetchFile fails")
	}
	if !store.IsDirty() {
		t.Fatal("expected store to remain dirty after error")
	}
}

func TestStartAndStop(t *testing.T) {
	store := NewStatusStore()
	mock := &mockGitClient{}

	rec := NewReconciler(store, mock, 10*time.Millisecond, "registry.yaml", "main")

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		rec.Start(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected Start to return after context cancellation")
	}
}
