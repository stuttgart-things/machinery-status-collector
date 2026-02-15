# Machinery Status Collector

Collects Crossplane claim status from multiple Kubernetes clusters and batches updates into pull requests against the central registry repository.

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Cluster A   │     │  Cluster B   │     │  Cluster N   │
│  (informer)  │     │  (informer)  │     │  (informer)  │
└──────┬───────┘     └──────┬───────┘     └──────┬───────┘
       │                    │                    │
       │  POST /api/v1/status                    │
       └────────────┬───────┴────────────────────┘
                    │
                    ▼
          ┌─────────────────┐
          │ Collector Server │
          │  (StatusStore)   │
          │  (Reconciler)    │
          └────────┬────────┘
                   │
                   │  GitHub API (PR)
                   ▼
          ┌─────────────────┐
          │ Registry Repo   │
          │ (YAML file)     │
          └─────────────────┘
```

### Components

- **Cluster Agent (informer)**: Runs inside each Kubernetes cluster. Uses a dynamic informer to watch Crossplane claim resources and POSTs status updates to the collector server.
- **Collector Server**: Central HTTP server that receives status updates, stores them in memory, and periodically reconciles them into pull requests against the registry repository.
- **Registry Repository**: GitHub repository containing the YAML registry file that tracks the status of all claims across clusters.

### Data Flow

1. The **informer** watches Crossplane claims for Add/Update events.
2. On each event, it extracts the Ready condition from `.status.conditions` and POSTs a JSON payload to the collector.
3. The **collector server** stores updates in a thread-safe in-memory store and marks it as dirty.
4. The **reconciler** periodically checks for dirty state, fetches the current registry file, updates claim statuses, and creates a pull request.

## Environment Variables

### Server Mode (`machinery-status-collector server`)

| Variable | Required | Default | Description |
|---|---|---|---|
| `GITHUB_TOKEN` | Yes | — | GitHub personal access token for API operations |
| `REGISTRY_REPO_OWNER` | Yes | — | GitHub repository owner |
| `REGISTRY_REPO_NAME` | Yes | — | GitHub repository name |
| `REGISTRY_FILE_PATH` | Yes | — | Path to the registry YAML file in the repo |
| `COLLECTOR_PORT` | No | `8095` | HTTP server listen port |
| `COLLECTOR_RECONCILE_INTERVAL` | No | `5m` | Reconciliation interval (Go duration) |
| `REGISTRY_BASE_BRANCH` | No | `main` | Base branch for pull requests |

### Informer Mode (`machinery-status-collector informer`)

| Variable | Required | Default | Description |
|---|---|---|---|
| `CLUSTER_NAME` | Yes | — | Name of the current cluster |
| `COLLECTOR_URL` | Yes | — | URL of the central collector API |
| `CLAIM_GROUP` | Yes | — | Crossplane claim API group |
| `CLAIM_VERSION` | No | `v1alpha1` | Crossplane claim API version |
| `CLAIM_RESOURCE` | Yes | — | Crossplane claim resource name |
| `CLAIM_NAMESPACE` | No | all | Namespace to watch (empty = all namespaces) |
| `KUBECONFIG` | No | `~/.kube/config` | Path to kubeconfig (ignored when running in-cluster) |

## Deployment

### Collector Server

Deploy the collector server as a Kubernetes Deployment with the required environment variables:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: machinery-status-collector
spec:
  replicas: 1
  selector:
    matchLabels:
      app: machinery-status-collector
  template:
    metadata:
      labels:
        app: machinery-status-collector
    spec:
      containers:
        - name: collector
          image: ghcr.io/stuttgart-things/machinery-status-collector:latest
          args: ["server"]
          ports:
            - containerPort: 8095
          env:
            - name: GITHUB_TOKEN
              valueFrom:
                secretKeyRef:
                  name: github-credentials
                  key: token
            - name: REGISTRY_REPO_OWNER
              value: stuttgart-things
            - name: REGISTRY_REPO_NAME
              value: machinery-registry
            - name: REGISTRY_FILE_PATH
              value: registry.yaml
```

### Cluster Agent (Informer)

Deploy the informer as a Deployment on each target cluster with a ServiceAccount that has read access to the Crossplane claim resources:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: machinery-status-informer
spec:
  replicas: 1
  selector:
    matchLabels:
      app: machinery-status-informer
  template:
    metadata:
      labels:
        app: machinery-status-informer
    spec:
      serviceAccountName: machinery-status-informer
      containers:
        - name: informer
          image: ghcr.io/stuttgart-things/machinery-status-collector:latest
          args: ["informer"]
          env:
            - name: CLUSTER_NAME
              value: cluster-01
            - name: COLLECTOR_URL
              value: http://machinery-status-collector.collector:8095
            - name: CLAIM_GROUP
              value: database.example.org
            - name: CLAIM_RESOURCE
              value: postgresqls
```

## Getting Started

### Prerequisites

- Go 1.25.6+
- [Task](https://taskfile.dev/) (optional)

### Installation

```bash
git clone https://github.com/stuttgart-things/machinery-status-collector.git
cd machinery-status-collector
go mod tidy
```

### Running

```bash
# Build with version info
task build

# Run the collector server
task run-server

# Run the cluster agent
task run-informer

# Run tests
task test
```

## API Reference

See the full [OpenAPI specification](openapi.yaml) for detailed request/response schemas.
