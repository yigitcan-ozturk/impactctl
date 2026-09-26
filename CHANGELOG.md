# Changelog

All notable changes to `impactctl` will be documented in this file.

The project follows semantic versioning from the first public release.

## [Unreleased] — v0.2 release candidate

### Added

- Optional explicit `.impactctl.yml` service-map configuration with service paths, criticality and ownership metadata.
- Evidence-backed OpenAPI provider/consumer service impact, conservative AsyncAPI event-contract classification and explicit dependency-aware downstream blast radius.
- Deterministic downstream service paths and human, JSON and Markdown service-impact output.
- Pinned real public multi-service repository dogfood validation covering catalog, orders, gateway and no-config regression; CI asserts repeat JSON determinism and catalog cross-format core semantic parity.

### Compatibility and boundaries

- Without a service map, the repository-level analysis remains available and does not invent a service dependency graph.
- Dependencies are declared, not automatically inferred. The core CLI remains local-first.
- The optional `oasdiff` semantic adapter (#12) is deferred beyond v0.2; no external analyzer is required for the core.
- External-repository validation is self-executed controlled-change evidence, not an independent practitioner review.
- v0.2.0 has **not** been tagged or released; v0.1.0 remains the current public install target.

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
