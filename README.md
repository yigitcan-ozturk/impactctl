# impactctl

**Know what your change can break — before you merge it.**

[![CI](https://github.com/yigitcan-ozturk/impactctl/actions/workflows/ci.yml/badge.svg)](https://github.com/yigitcan-ozturk/impactctl/actions/workflows/ci.yml)

`impactctl` is a deterministic change-impact CLI for pull requests. It turns a Git diff into a compact risk signal by inspecting technical contracts, database migrations, deployment/configuration changes, CI/CD files and repository ownership.

> Early-stage open source. The v0.1 goal is intentionally narrow: make change impact visible in seconds, without requiring a hosted service or AI.

## Why

A small code diff can have a large system impact. Reviewers often see the changed lines but miss the surrounding blast radius: API contracts, migrations, deployment files, runtime configuration and ownership boundaries.

`impactctl` answers one question:

**What deserves extra attention before this change is merged?**

## Example

```text
$ impactctl pr --base main --head HEAD

CRITICAL IMPACT  (75/100)
────────────────────────────────────────────
Changed files          7
Findings               3
Owner teams            2

Why
! contract     api/openapi.yaml changes an API contract
! database     db/migrations/024_vendor.sql looks like a database migration
! ownership    change crosses multiple ownership areas

Suggested review
→ @platform-team
→ @procurement-team
```

## v0.1 signals

- Git diff scope
- OpenAPI / Swagger contract changes
- Database migrations
- Terraform / Kubernetes / Helm / Docker changes
- GitHub Actions and common CI changes
- Runtime/configuration changes
- `CODEOWNERS` review boundaries
- Deterministic `LOW / MEDIUM / HIGH / CRITICAL` risk classification
- Human-readable and JSON output

## Install from source

```bash
go install github.com/yigitcan-ozturk/impactctl/cmd/impactctl@latest
```

Until the first release is tagged, clone and build locally:

```bash
git clone https://github.com/yigitcan-ozturk/impactctl.git
cd impactctl
go build -o impactctl ./cmd/impactctl
./impactctl pr --base main --head HEAD
```

## Design principles

- **Deterministic core** — the same repository state should produce the same result.
- **Evidence before confidence** — every risk signal should be explainable.
- **Local-first** — source code does not need to leave the developer environment.
- **Fast adoption** — a single CLI should provide value before configuration is required.
- **Extensible, not monolithic** — language and platform adapters can grow around a small core.

## Roadmap

### v0.1 — repository impact
- [x] Git diff
- [x] contract / migration / deployment / CI / config signals
- [x] CODEOWNERS-aware review hints
- [x] JSON output
- [ ] GitHub Action PR comment
- [ ] golden-fixture integration tests

### v0.2 — service impact
- [ ] service map configuration
- [ ] API consumer/provider relationships
- [ ] event schema changes
- [ ] dependency-aware downstream impact

### v0.3 — system impact
- [ ] Kubernetes workload relationships
- [ ] Terraform resource ownership
- [ ] cross-repository impact contracts
- [ ] architecture boundary policies

## Contributing

Contributions are welcome. See [`CONTRIBUTING.md`](CONTRIBUTING.md) and the issues labeled `good first issue` or `help wanted` for scoped starting points.

## Status

`impactctl` is being built in public. The first milestone is a small, trusted CLI that developers can run on any repository in seconds.

Contributions, edge cases and real-world examples are welcome.

## License

MIT
