package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/stuttgart-things/machinery-status-collector/internal/informer"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var informerCmd = &cobra.Command{
	Use:   "informer",
	Short: "Start the Kubernetes claim watcher",
	Long: `Start the cluster agent that watches Crossplane claim resources via a
Kubernetes dynamic informer and POSTs status updates to the central
collector API.`,
	RunE: runInformer,
}

func init() {
	rootCmd.AddCommand(informerCmd)
}

func runInformer(cmd *cobra.Command, args []string) error {
	// Required environment variables.
	required := []string{
		"CLUSTER_NAME",
		"COLLECTOR_URL",
		"CLAIM_GROUP",
		"CLAIM_RESOURCE",
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

	clusterName := os.Getenv("CLUSTER_NAME")
	collectorURL := os.Getenv("COLLECTOR_URL")
	claimGroup := os.Getenv("CLAIM_GROUP")
	claimResource := os.Getenv("CLAIM_RESOURCE")

	claimVersion := os.Getenv("CLAIM_VERSION")
	if claimVersion == "" {
		claimVersion = "v1alpha1"
	}

	claimNamespace := os.Getenv("CLAIM_NAMESPACE")

	// Build Kubernetes client (in-cluster or kubeconfig).
	dynamicClient, err := buildDynamicClient()
	if err != nil {
		return fmt.Errorf("build kubernetes client: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    claimGroup,
		Version:  claimVersion,
		Resource: claimResource,
	}

	watcher := informer.NewClaimWatcher(dynamicClient, collectorURL, clusterName, gvr, claimNamespace)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("informer starting: cluster=%s gvr=%s/%s/%s namespace=%q",
			clusterName, claimGroup, claimVersion, claimResource, claimNamespace)
		if err := watcher.Start(ctx); err != nil && ctx.Err() == nil {
			errCh <- err
		}
	}()

	// Wait for signal or watcher error.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("received signal %v, shutting down", sig)
	case err := <-errCh:
		return fmt.Errorf("informer error: %w", err)
	}

	cancel()
	log.Println("informer stopped")
	return nil
}

func buildDynamicClient() (dynamic.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig.
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig: %w", err)
		}
	}
	return dynamic.NewForConfig(config)
}
