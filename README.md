# machinery-status-collector

[![Release](https://github.com/stuttgart-things/machinery-status-collector/actions/workflows/release.yml/badge.svg)](https://github.com/stuttgart-things/machinery-status-collector/actions/workflows/release.yml)

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

## Prerequisites

- Go 1.25.6+
- [Task](https://taskfile.dev/) (optional, for task automation)

## Getting Started

```bash
git clone https://github.com/stuttgart-things/machinery-status-collector.git
cd machinery-status-collector
go mod tidy

# Build
task build

# Run tests
task test
```

## Server

Start the HTTP server and reconciler that collects status updates and creates pull requests:

```bash
task run-server
```

### Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `GITHUB_TOKEN` | Yes | — | GitHub personal access token |
| `REGISTRY_REPO_OWNER` | Yes | — | GitHub repository owner |
| `REGISTRY_REPO_NAME` | Yes | — | GitHub repository name |
| `REGISTRY_FILE_PATH` | Yes | — | Path to registry YAML in repo |
| `COLLECTOR_PORT` | No | `8095` | HTTP listen port |
| `COLLECTOR_RECONCILE_INTERVAL` | No | `5m` | Reconcile ticker interval |
| `REGISTRY_BASE_BRANCH` | No | `main` | Base branch for PRs |

### Example

```bash
export GITHUB_TOKEN=ghp_...
export REGISTRY_REPO_OWNER=stuttgart-things
export REGISTRY_REPO_NAME=my-registry
export REGISTRY_FILE_PATH=clusters/registry.yaml

task run-server
```

## Informer (Cluster Agent)

Start the Kubernetes dynamic informer that watches Crossplane claim resources and POSTs status updates to the collector server:

```bash
task run-informer
```

### Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `CLUSTER_NAME` | Yes | — | Name of the current cluster |
| `COLLECTOR_URL` | Yes | — | URL of the central collector API |
| `CLAIM_GROUP` | Yes | — | Crossplane claim API group |
| `CLAIM_VERSION` | No | `v1alpha1` | Crossplane claim API version |
| `CLAIM_RESOURCE` | Yes | — | Crossplane claim resource name |
| `CLAIM_NAMESPACE` | No | all | Namespace to watch (empty = all namespaces) |
| `KUBECONFIG` | No | `~/.kube/config` | Path to kubeconfig (ignored in-cluster) |

### Example

```bash
export CLUSTER_NAME=cluster-01
export COLLECTOR_URL=http://localhost:8095
export CLAIM_GROUP=database.example.org
export CLAIM_RESOURCE=postgresqls

task run-informer
```

## API Usage

### Submit a status update

```bash
curl -X POST http://localhost:8095/api/v1/status \
  -H "Content-Type: application/json" \
  -d '{"cluster":"cluster-a","claimRef":"network/vpc-prod","statusMessage":"ready"}'
```

### Get all status entries

```bash
curl http://localhost:8095/api/v1/status
```

### Get status entries for a specific cluster

```bash
curl http://localhost:8095/api/v1/status/cluster-a
```

### Health check

```bash
curl http://localhost:8095/healthz
# {"status":"ok"}
```

### Version info

```bash
curl http://localhost:8095/version
# {"version":"v0.1.0","commit":"abc1234"}
```

See the full [OpenAPI specification](docs/openapi.yaml) for detailed request/response schemas.

## Development

| Task | Description |
|------|-------------|
| `task build` | Build binary with version info |
| `task test` | Run tests with race detection and coverage |
| `task lint` | Run golangci-lint |
| `task run-server` | Build and run the server |
| `task run-informer` | Build and run the informer |
| `task tag -- v0.1.0` | Create and push a git tag |

## Releases

Automated releases via GitHub Actions with [semantic-release](https://github.com/semantic-release/semantic-release).

Please use the [Conventional Commits](https://www.conventionalcommits.org/) format:
- `feat: add new API for X`
- `fix: resolve memory leak`
- `chore: update dependencies`

## License

See [LICENSE](LICENSE) for details.
