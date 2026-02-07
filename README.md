# machinery-status-collector

[![Release](https://github.com/stuttgart-things/machinery-status-collector/actions/workflows/release.yml/badge.svg)](https://github.com/stuttgart-things/machinery-status-collector/actions/workflows/release.yml)

Monitor deployed claims and report health status

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

## Releases

- Automatisierte Releases über GitHub Actions mit [semantic-release](https://github.com/semantic-release/semantic-release).
- Konfiguration: `.releaserc.json`, Workflow: `.github/workflows/release.yml`, Changelog: `CHANGELOG.md`.
- Branches: `main` (Stable), `release/next` (Release-Branch für Changelog-Push).

### Konventionelle Commits

Bitte nutze das [Conventional Commits](https://www.conventionalcommits.org/) Format, z. B.:
- `feat: Neue API für X`
- `fix: Behebt Speicherleck`
- `chore: Abhängigkeiten aktualisiert`

### Lokaler Dry-Run (optional)

Falls Node.js installiert ist, kann ein Dry-Run getestet werden:

```bash
npx semantic-release --dry-run
```

## Branchschutz (empfohlen)

Richte Branch-Protection in GitHub ein, um stabile Releases sicherzustellen:

- **Geschützte Branches**: `main` und `release/next`
- **Require pull request reviews before merging**: aktivieren (mind. 1 Review)
- **Require status checks to pass before merging**: aktivieren und den Workflow `Release` auswählen
- **Require linear history**: optional aktivieren
- **Restrict who can push to matching branches**: optional (nur Maintainer)
- **Dismiss stale pull request approvals when new commits are pushed**: optional aktivieren

Hinweis: Diese Einstellungen findest du unter
GitHub → Repository → Settings → Branches → Branch protection rules.

### Alternative: CLI (gh)

Mit der GitHub CLI kann Branchschutz gesetzt werden (Admin-Rechte erforderlich):

```bash
# Werte anpassen
OWNER=stuttgart-things
REPO=machinery-status-collector

# Schutz für main
gh api -X PUT \
	repos/$OWNER/$REPO/branches/main/protection \
	-H "Accept: application/vnd.github+json" \
	-F required_status_checks.strict=true \
	-F required_status_checks.contexts='["Release"]' \
	-F enforce_admins=true \
	-F required_pull_request_reviews.dismiss_stale_reviews=true \
	-F required_pull_request_reviews.required_approving_review_count=1 \
	-F restrictions=null

# Schutz für release/next
gh api -X PUT \
	repos/$OWNER/$REPO/branches/release/next/protection \
	-H "Accept: application/vnd.github+json" \
	-F required_status_checks.strict=true \
	-F required_status_checks.contexts='["Release"]' \
	-F enforce_admins=true \
	-F required_pull_request_reviews.dismiss_stale_reviews=true \
	-F required_pull_request_reviews.required_approving_review_count=1 \
	-F restrictions=null
```

Hinweis: `gh` nutzt deine lokale Authentifizierung (`gh auth login`).

## License

See [LICENSE](LICENSE) for details.
