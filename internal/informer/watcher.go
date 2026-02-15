package informer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

// ClaimWatcher watches Crossplane claim resources via a dynamic informer and
// POSTs status updates to the central collector API.
type ClaimWatcher struct {
	dynamicClient dynamic.Interface
	collectorURL  string
	clusterName   string
	gvr           schema.GroupVersionResource
	namespace     string
	httpClient    *http.Client
}

// NewClaimWatcher creates a ClaimWatcher for the given Crossplane claim GVR.
func NewClaimWatcher(dynamicClient dynamic.Interface, collectorURL, clusterName string, gvr schema.GroupVersionResource, namespace string) *ClaimWatcher {
	return &ClaimWatcher{
		dynamicClient: dynamicClient,
		collectorURL:  collectorURL,
		clusterName:   clusterName,
		gvr:           gvr,
		namespace:     namespace,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Start begins watching the configured GVR and blocks until ctx is cancelled.
func (w *ClaimWatcher) Start(ctx context.Context) error {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		w.dynamicClient, 0, w.namespace, nil,
	)

	informer := factory.ForResource(w.gvr).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u, ok := obj.(*unstructured.Unstructured)
			if !ok {
				return
			}
			if err := w.sendStatus(u); err != nil {
				slog.Error("send status on add", "error", err)
			}
		},
		UpdateFunc: func(_, newObj interface{}) {
			u, ok := newObj.(*unstructured.Unstructured)
			if !ok {
				return
			}
			if err := w.sendStatus(u); err != nil {
				slog.Error("send status on update", "error", err)
			}
		},
	})

	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())

	<-ctx.Done()
	return ctx.Err()
}

type statusPayload struct {
	Cluster       string `json:"cluster"`
	ClaimRef      string `json:"claimRef"`
	StatusMessage string `json:"statusMessage"`
}

// sendStatus extracts the claim status and POSTs it to the collector API.
func (w *ClaimWatcher) sendStatus(claim *unstructured.Unstructured) error {
	statusMsg, err := ExtractClaimStatus(claim)
	if err != nil {
		return fmt.Errorf("extract status: %w", err)
	}

	claimRef := fmt.Sprintf("%s/%s", claim.GetNamespace(), claim.GetName())

	payload := statusPayload{
		Cluster:       w.clusterName,
		ClaimRef:      claimRef,
		StatusMessage: statusMsg,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/status", w.collectorURL)
	resp, err := w.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("post status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	slog.Info("status sent", "cluster", w.clusterName, "claimRef", claimRef, "status", statusMsg)
	return nil
}
