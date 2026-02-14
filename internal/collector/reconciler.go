package collector

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/stuttgart-things/machinery-status-collector/internal/registry"
)

// GitClient abstracts the GitHub operations needed by the Reconciler.
type GitClient interface {
	FetchFile(path, ref string) ([]byte, string, error)
	CreateBranch(baseSHA, branchName string) error
	UpdateFile(path, branchName, message string, content []byte, sha string) error
	CreatePR(title, body, head, base string) (int, error)
	ListOpenPRs(head string) ([]int, error)
	GetRef(branch string) (string, error)
}

// Reconciler periodically checks the status store for dirty entries and
// opens a PR with the updated registry YAML.
type Reconciler struct {
	store        *StatusStore
	gitClient    GitClient
	interval     time.Duration
	registryPath string
	baseBranch   string
}

// NewReconciler creates a Reconciler that checks the store at the given interval.
func NewReconciler(store *StatusStore, gitClient GitClient, interval time.Duration, registryPath, baseBranch string) *Reconciler {
	return &Reconciler{
		store:        store,
		gitClient:    gitClient,
		interval:     interval,
		registryPath: registryPath,
		baseBranch:   baseBranch,
	}
}

// Start runs the reconciliation loop until the context is cancelled.
func (r *Reconciler) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.reconcileOnce(ctx); err != nil {
				log.Printf("reconcile error: %v", err)
			}
		}
	}
}

func (r *Reconciler) reconcileOnce(_ context.Context) error {
	if !r.store.IsDirty() {
		return nil
	}

	yamlBytes, fileSHA, err := r.gitClient.FetchFile(r.registryPath, r.baseBranch)
	if err != nil {
		return fmt.Errorf("fetch registry: %w", err)
	}

	reg, err := registry.ParseRegistry(yamlBytes)
	if err != nil {
		return fmt.Errorf("parse registry: %w", err)
	}

	for _, entry := range r.store.GetAll() {
		registry.UpdateClaimStatus(reg, entry.Cluster, entry.ClaimRef, entry.StatusMessage)
	}

	updatedYAML, err := registry.SerializeRegistry(reg)
	if err != nil {
		return fmt.Errorf("serialize registry: %w", err)
	}

	branchName := fmt.Sprintf("status-update-%d", time.Now().Unix())

	openPRs, err := r.gitClient.ListOpenPRs(branchName)
	if err != nil {
		return fmt.Errorf("list open PRs: %w", err)
	}
	if len(openPRs) > 0 {
		return nil
	}

	commitSHA, err := r.gitClient.GetRef(r.baseBranch)
	if err != nil {
		return fmt.Errorf("get ref: %w", err)
	}

	if err := r.gitClient.CreateBranch(commitSHA, branchName); err != nil {
		return fmt.Errorf("create branch: %w", err)
	}

	if err := r.gitClient.UpdateFile(r.registryPath, branchName, "chore: update claim statuses", updatedYAML, fileSHA); err != nil {
		return fmt.Errorf("update file: %w", err)
	}

	prNum, err := r.gitClient.CreatePR(
		"chore: update claim statuses",
		"Automated status update from machinery-status-collector.",
		branchName,
		r.baseBranch,
	)
	if err != nil {
		return fmt.Errorf("create PR: %w", err)
	}

	log.Printf("created PR #%d on branch %s", prNum, branchName)
	r.store.MarkFlushed()
	return nil
}
