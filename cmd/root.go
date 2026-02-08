package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version information (set via ldflags during build)
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "machinery-status-collector",
	Short: "Machinery Status Collector - Collect and reconcile Crossplane claim status",
	Long: `Machinery Status Collector collects Crossplane claim status from multiple
Kubernetes clusters and batches updates into pull requests against the
central registry repository.

By default, it starts the collector server. Use subcommands for other operations.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
