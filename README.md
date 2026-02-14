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
go run .
# or with Task
task run
```

## API Usage

The collector exposes an HTTP API for receiving and querying status updates.

### Submit a status update

```bash
curl -X POST http://localhost:8080/api/v1/status \
  -H "Content-Type: application/json" \
  -d '{"cluster":"cluster-a","claimRef":"network/vpc-prod","statusMessage":"ready"}'
```

### Get all status entries

```bash
curl http://localhost:8080/api/v1/status
```

### Get status entries for a specific cluster

```bash
curl http://localhost:8080/api/v1/status/cluster-a
```

### Health check

```bash
curl http://localhost:8080/healthz
# {"status":"ok"}
```

### Version info

```bash
curl http://localhost:8080/version
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
