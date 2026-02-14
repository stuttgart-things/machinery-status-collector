# machinery-status-collector

[![Release](https://github.com/stuttgart-things/machinery-status-collector/actions/workflows/release.yml/badge.svg)](https://github.com/stuttgart-things/machinery-status-collector/actions/workflows/release.yml)

Collects Crossplane claim status from multiple Kubernetes clusters and batches updates into pull requests against the central registry repository.

## Prerequisites

- Go 1.25.6+
- [Task](https://taskfile.dev/) (optional, for task automation)

## Getting Started

```bash
# Clone the repository
git clone https://github.com/stuttgart-things/machinery-status-collector.git
cd machinery-status-collector

# Install dependencies
go mod tidy

# Run the application
go run . server
```

## Server

Start the HTTP server and reconciler:

```bash
go run . server
```

### Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `GITHUB_TOKEN` | yes | — | GitHub API token |
| `REGISTRY_REPO_OWNER` | yes | — | GitHub org/user |
| `REGISTRY_REPO_NAME` | yes | — | GitHub repo name |
| `REGISTRY_FILE_PATH` | yes | — | Path to registry YAML in repo |
| `COLLECTOR_PORT` | no | `8095` | HTTP listen port |
| `COLLECTOR_RECONCILE_INTERVAL` | no | `5m` | Reconcile ticker interval |
| `REGISTRY_BASE_BRANCH` | no | `main` | Base branch for PRs |

### Example

```bash
export GITHUB_TOKEN=ghp_...
export REGISTRY_REPO_OWNER=stuttgart-things
export REGISTRY_REPO_NAME=my-registry
export REGISTRY_FILE_PATH=clusters/registry.yaml

go run . server
```

The server listens on `:8095` by default. The reconciler runs in the background, periodically batching collected status updates into pull requests. Graceful shutdown is triggered by `SIGINT` or `SIGTERM`.

## API Usage

The collector exposes an HTTP API for receiving and querying status updates.

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
# {"version":"dev","commit":"none"}
```

## Releases

Automated releases via GitHub Actions with [semantic-release](https://github.com/semantic-release/semantic-release).

Configuration: `.releaserc.json`, Workflow: `.github/workflows/release.yml`, Changelog: `CHANGELOG.md`.

Branches: `main` (stable), `release/next` (release branch for changelog push).

### Conventional Commits

Please use the [Conventional Commits](https://www.conventionalcommits.org/) format, e.g.:
- `feat: add new API for X`
- `fix: resolve memory leak`
- `chore: update dependencies`

## License

See [LICENSE](LICENSE) for details.
