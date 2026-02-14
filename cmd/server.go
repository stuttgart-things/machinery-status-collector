package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/stuttgart-things/machinery-status-collector/internal/api"
	"github.com/stuttgart-things/machinery-status-collector/internal/collector"
	"github.com/stuttgart-things/machinery-status-collector/internal/git"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the collector server",
	Long: `Start the machinery-status-collector HTTP server and reconciler.

The server accepts status updates via REST API and periodically reconciles
them into pull requests against the central registry repository.`,
	RunE: runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	// Required environment variables.
	required := []string{
		"GITHUB_TOKEN",
		"REGISTRY_REPO_OWNER",
		"REGISTRY_REPO_NAME",
		"REGISTRY_FILE_PATH",
	}
	var missing []string
	for _, key := range required {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables:\n  %s", strings.Join(missing, "\n  "))
	}

	token := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("REGISTRY_REPO_OWNER")
	repo := os.Getenv("REGISTRY_REPO_NAME")
	filePath := os.Getenv("REGISTRY_FILE_PATH")

	// Optional environment variables with defaults.
	port := os.Getenv("COLLECTOR_PORT")
	if port == "" {
		port = "8095"
	}

	intervalStr := os.Getenv("COLLECTOR_RECONCILE_INTERVAL")
	if intervalStr == "" {
		intervalStr = "5m"
	}
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return fmt.Errorf("invalid COLLECTOR_RECONCILE_INTERVAL: %w", err)
	}

	baseBranch := os.Getenv("REGISTRY_BASE_BRANCH")
	if baseBranch == "" {
		baseBranch = "main"
	}

	// Create dependencies.
	store := collector.NewStatusStore()
	gitClient := git.NewGitHubClient(token, owner, repo)
	rec := collector.NewReconciler(store, gitClient, interval, filePath, baseBranch)
	apiServer := api.NewServer(store, Version, Commit)

	// Start reconciler in background.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go rec.Start(ctx)

	// Start HTTP server.
	addr := ":" + port
	srv := &http.Server{Addr: addr, Handler: apiServer.Handler}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for signal or server error.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("received signal %v, shutting down", sig)
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	// Graceful shutdown.
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	log.Println("server stopped")
	return nil
}
