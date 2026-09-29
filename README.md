# impactctl

**Know what your change can break — before you merge it.**

[![CI](https://github.com/yigitcan-ozturk/impactctl/actions/workflows/ci.yml/badge.svg)](https://github.com/yigitcan-ozturk/impactctl/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/yigitcan-ozturk/impactctl)](https://github.com/yigitcan-ozturk/impactctl/releases/latest)

`impactctl` is a deterministic change-impact CLI for pull requests. It turns a Git diff into a compact risk signal by inspecting technical contracts, database migrations, deployment/configuration changes, CI/CD files and repository ownership.

> `v0.2.0` is the current public release. It extends the original repository-level change signals into explicit service relationships, API/event-contract evidence and deterministic downstream system impact.

## Proof at a glance

| Signal | Evidence |
| --- | --- |
| **Current release** | **v0.2.0** |
| **Core model** | Deterministic, local-first change-impact analysis |
| **System context** | Explicit service maps, API/event contracts and downstream dependency paths |
| **Validation** | Pinned external-repository candidate with controlled catalog, orders, gateway and no-config scenarios |
| **CI integration** | Live-validated GitHub PR comment workflow |
| **Release engineering** | Cross-platform binaries, version injection, checksums and CI smoke coverage |

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

## Released v0.2 capabilities

- Git diff scope
- OpenAPI / Swagger contract changes
- Database migrations
- Terraform / Kubernetes / Helm / Docker changes
- GitHub Actions and common CI changes
- Runtime/configuration changes
- `CODEOWNERS` review boundaries from `.github/CODEOWNERS`, `CODEOWNERS` or `docs/CODEOWNERS`
- Deterministic `LOW / MEDIUM / HIGH / CRITICAL` risk classification
- Human-readable, JSON and GitHub-flavored Markdown output

## Install

Install the released Go CLI:

```bash
go install github.com/yigitcan-ozturk/impactctl/cmd/impactctl@v0.2.0
```

Or download a prebuilt binary for Linux, macOS or Windows from the [latest GitHub release](https://github.com/yigitcan-ozturk/impactctl/releases/latest). Release archives include SHA-256 checksums.

Build from source:

```bash
git clone https://github.com/yigitcan-ozturk/impactctl.git
cd impactctl
go build -o impactctl ./cmd/impactctl
./impactctl pr --base main --head HEAD
```

For machine-readable output:

```bash
./impactctl pr --base main --head HEAD --json
```

For a pull-request comment payload:

```bash
./impactctl pr --base main --head HEAD --markdown
```

## v0.2 service impact

The v0.2.0 release supports an **optional, explicit** `.impactctl.yml` service map. This enables path-to-service matching, configured OpenAPI provider/consumer impact, conservative AsyncAPI evidence and deterministic downstream dependency paths. It does not discover runtime dependencies automatically. Without the file, the CLI continues its existing repository-level analysis.

Minimal example:

```yaml
version: 1
services:
  - name: catalog
    paths: [services/catalog/**]
    criticality: high
  - name: orders
    paths: [services/orders/**]
    criticality: high
    depends_on: [catalog]
  - name: gateway
    paths: [services/gateway/**]
    criticality: medium
    depends_on: [orders, catalog]
```

Run `impactctl pr --base main --head HEAD` as before. Use `--json` or `--markdown` for machine-readable and pull-request formats. A catalog change in the example directly affects catalog and identifies orders and gateway as declared downstream dependents; a gateway-only change has no declared downstream service. Review [the full service-map schema](docs/SERVICE_MAP.md) before adding contract mappings or owner metadata.

**Upgrade path:** Existing v0.1 users need not add configuration. To enable v0.2 service impact, explicitly document service paths and real dependency edges, review the resulting impact against repository code, and add optional contract metadata as appropriate. Do not treat absent edges as discovered facts. The optional `oasdiff` adapter is deferred beyond v0.2; no network service is required for the core CLI.

The [pinned external-repository validation](docs/validation/V0.2_EXTERNAL_CANDIDATE.md) passed controlled catalog, orders, gateway and no-config scenarios in CI. This is reproducible self-run dogfood evidence, **not** an independent practitioner endorsement. The public install command above is pinned to the released v0.2.0 baseline.

## Experimental beyond v0.2: SAP landscape impact

`impactctl` now includes an experimental, local-first SAP/enterprise landscape spike that asks a different question:

**If this explicit SAP component changes, what surrounding integrations, applications and business processes deserve review — and through which dependency path?**

Run the synthetic credential-free example:

```bash
go run ./cmd/impactctl sap --manifest examples/sap-vendor-status.yml
```

Example impact path:

```text
Z_VENDOR_STATUS
  -> BTP_VENDOR_IFLOW
  -> SUPPLIER_PORTAL
  -> VENDOR_APPROVAL
```

The spike uses only explicit dependencies from a strict local YAML manifest. It does not connect to SAP systems, infer unknown relationships or replace SAP-native change controls. See [`docs/SAP_IMPACT.md`](docs/SAP_IMPACT.md) for schema, scoring and product boundaries.

## Try it on a real repository

The most useful next validation is simple: run `impactctl` against a real pull request and check whether the reported review scope matches what an experienced maintainer would expect.

Useful feedback includes:

- a change that should have been higher or lower risk;
- a false-positive config, migration, contract or ownership signal;
- a service relationship that should or should not propagate downstream;
- a repository pattern that the current service-map model cannot express cleanly;
- a PR where the suggested review scope changed a real review decision.

Reproducible examples are especially valuable. Open an issue with the repository structure, relevant changed paths, expected result and actual `impactctl` output. Sanitized or synthetic reproductions are welcome when the original repository cannot be shared.

## GitHub pull-request comments

The included `PR Impact` workflow runs `impactctl` on same-repository pull requests and creates one bot comment containing the current risk level, findings and suggested owners. New pushes update the existing marked comment instead of creating duplicates.

The v0.1 workflow intentionally does **not** execute fork-supplied code with a write-capable token. External-fork comment support can be added later with a separated trusted workflow pattern.

## Design principles

- **Deterministic core** — the same repository state should produce the same result.
- **Evidence before confidence** — every risk signal should be explainable.
- **Local-first** — source code does not need to leave the developer environment.
- **Fast adoption** — a single CLI should provide value before configuration is required.
- **Extensible, not monolithic** — language and platform adapters can grow around a small core.
- **Safe CI integration** — untrusted fork code should not receive write-capable workflow credentials.

## Roadmap

### v0.1 — repository impact
- [x] Git diff
- [x] contract / migration / deployment / CI / config signals
- [x] CODEOWNERS-aware review hints
- [x] JSON output
- [x] Markdown PR payload
- [x] golden-fixture integration tests
- [x] GitHub Action PR comment validated on a live pull request
- [x] cross-platform `v0.1.0` release with checksums

### v0.2 — service impact
- [x] [service-map configuration](https://github.com/yigitcan-ozturk/impactctl/issues/8)
- [x] [API consumer/provider relationships](https://github.com/yigitcan-ozturk/impactctl/issues/9)
- [x] [AsyncAPI event-schema impact](https://github.com/yigitcan-ozturk/impactctl/issues/10)
- [x] [dependency-aware downstream impact](https://github.com/yigitcan-ozturk/impactctl/issues/11)
- [ ] [optional `oasdiff` semantic adapter](https://github.com/yigitcan-ozturk/impactctl/issues/12)
- [x] [real multi-service repository validation](https://github.com/yigitcan-ozturk/impactctl/issues/15)

The v0.2 direction is deliberately composable: `impactctl` should own **change → service/system impact** while interoperating with specialist analyzers where they already provide deeper domain semantics.

The v0.2 [service-map schema](docs/SERVICE_MAP.md) covers path-to-service mapping, direct OpenAPI provider/consumer relationships and explicit dependency-aware downstream paths. [AsyncAPI impact analysis](docs/ASYNCAPI.md) adds conservative `ADDITIVE / BREAKING / REVIEW` event-contract evidence. See the [v0.2 execution plan and release gate](docs/V0.2_EXECUTION.md) for sequencing and acceptance criteria. Public adoption signals are tracked separately in [traction](docs/TRACTION.md).

### v0.3 — system impact
- [x] experimental SAP/enterprise landscape manifest spike
- [ ] Kubernetes workload relationships
- [ ] Terraform resource ownership
- [ ] cross-repository impact contracts
- [ ] architecture boundary policies
- [ ] evidence-backed enterprise adapters for selected external systems

## Contributing

Contributions are welcome. See [`CONTRIBUTING.md`](CONTRIBUTING.md) and the issues labeled `good first issue` or `help wanted` for scoped starting points.

## Status

`impactctl v0.2.0` is publicly released and being built in the open. The current baseline combines repository-level change signals with explicit service relationships and deterministic downstream impact while remaining local-first and explainable.

The PR comment workflow was live-validated on the repository's own integration path. Release packaging, version injection and checksum generation are covered by CI smoke tests.

Experimental enterprise adapters remain additive and do not change the released v0.2 baseline.

Contributions, edge cases and real-world examples are welcome.

## License

MIT