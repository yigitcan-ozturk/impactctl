# Changelog

All notable changes to `impactctl` will be documented in this file.

The project follows semantic versioning from the first public release.

## [0.1.0] - 2026-08-30

### Added

- Deterministic pull-request change-impact analysis from Git diffs.
- `LOW`, `MEDIUM`, `HIGH`, and `CRITICAL` risk classification with an explainable score.
- OpenAPI and Swagger contract-change detection.
- Database migration detection.
- Terraform, Kubernetes, Helm, Docker and deployment/infrastructure change detection.
- GitHub Actions, GitLab CI and Jenkins CI/CD change detection.
- Runtime/configuration change detection.
- `CODEOWNERS`-aware review hints with standard GitHub lookup precedence:
  - `.github/CODEOWNERS`
  - `CODEOWNERS`
  - `docs/CODEOWNERS`
- Human-readable terminal output.
- Machine-readable JSON output via `impactctl pr --json`.
- GitHub-flavored Markdown output via `impactctl pr --markdown`.
- Safe same-repository GitHub Actions workflow that creates or updates one marked PR impact comment.
- End-to-end golden fixture proving a deterministic `CRITICAL` result across API, database, deployment and ownership boundaries.
- CI gates for `go test ./...`, `go vet ./...` and `go build ./cmd/impactctl`.
- MIT license and contributor guidance.

### Security / workflow boundary

The v0.1 PR-comment workflow does not run fork-supplied code with a write-capable token. External-fork comment support is intentionally deferred until it can be implemented with a separated trusted workflow design.

[0.1.0]: https://github.com/yigitcan-ozturk/impactctl/releases/tag/v0.1.0
