# Service map configuration

`impactctl` v0.2 introduces an optional repository-local service map at `.impactctl.yml`.

The file gives the analyzer explicit evidence for mapping changed repository paths to logical services, declaring OpenAPI provider/consumer relationships, and describing service dependency edges. It is optional: if the file does not exist, v0.1 repository-level analysis continues unchanged.

## Schema version

The first schema version is `1`.

```yaml
version: 1
services:
  - name: orders
    paths:
      - services/orders/**
    criticality: high
    owners:
      - "@orders-team"
    openapi:
      - path: api/orders/openapi.yaml
        consumers:
          - checkout

  - name: checkout
    paths:
      - services/checkout/**
    criticality: medium
    owners:
      - "@checkout-team"
    depends_on:
      - orders

  - name: billing
    paths:
      - services/billing/**
    criticality: critical
    owners:
      - "@billing-team"
    depends_on:
      - checkout
```

In this example:

- `orders` provides the OpenAPI contract.
- `checkout` is an explicitly declared consumer of that contract and also explicitly depends on `orders`.
- `billing` depends on `checkout`.
- A change to `orders` can therefore surface `orders → checkout` and `orders → checkout → billing` as explainable downstream paths.

## Fields

### `version`

Required integer. v0.2 starts with schema version `1`. Unsupported versions fail clearly instead of being interpreted silently.

### `services`

Required non-empty list of service definitions.

Each service supports:

- `name` — required unique logical service name
- `paths` — required non-empty list of repository-relative path patterns
- `criticality` — optional: `low`, `medium`, `high`, or `critical`
- `owners` — optional list of team/user identifiers
- `depends_on` — optional list of logical upstream service names
- `openapi` — optional list of OpenAPI contracts provided by this service

`depends_on` is directional. If `checkout` declares `depends_on: [orders]`, a change to `orders` may affect `checkout` downstream. Dependency names must reference another declared service. Empty, unknown, duplicate and self-dependencies are rejected.

Each `openapi` entry supports:

- `path` — required repository-relative contract path or supported path pattern
- `consumers` — optional list of service names that explicitly consume the contract

Every consumer must name another service declared in the same file. Unknown consumers, duplicate consumers and self-consumption are rejected.

Unknown fields are rejected so configuration mistakes do not silently change analysis behavior.

## Path matching

Repository paths always use forward slashes.

Supported forms:

- directory prefix: `services/checkout`
- recursive directory prefix: `services/checkout/**`
- Go-style path glob: `api/*.yaml`
- exact file path: `api/openapi.yaml`

Absolute paths and patterns that escape the repository with `..` are rejected.

A changed path may belong to more than one configured service. Matching services are returned in deterministic name order.

OpenAPI contract matching uses the same path rules. The changed file path is preserved in the impact report as evidence.

## Explicit relationship rule

`impactctl` does not infer service relationships from imports, naming, network calls, repository proximity or other heuristics.

Relationships enter the graph only through explicit configuration:

- `depends_on` creates a general service dependency edge.
- an OpenAPI provider plus an explicit `consumers` declaration creates a provider → consumer impact edge.

If an OpenAPI file changes and no provider/consumer mapping is declared, the normal repository-level contract finding still appears, but no service relationship is fabricated.

## Downstream impact

Changed files are first mapped to configured services. Those directly changed services become traversal sources.

`impactctl` then walks explicit downstream edges and reports each affected service with an evidence path. The traversal is:

- deterministic
- cycle-safe
- multi-source
- local-only
- bounded by the configured graph

Linear example:

```text
orders → checkout → billing
```

Branching graphs report each reachable service. Cycles terminate safely and do not re-add the source as its own downstream impact.

When the same downstream service can be reached through more than one path, the deterministic traversal keeps one stable path rather than emitting duplicate service impacts.

## Risk contribution

Direct OpenAPI provider/consumer relationships remain evidence and do not add a separate score by themselves.

For v0.2 downstream analysis, each reachable service configured with `criticality: critical` contributes a **+20 downstream risk finding**. The finding includes the dependency path that caused the escalation. Existing total risk remains capped at 100.

`low`, `medium`, and `high` downstream criticalities are currently reported as metadata but do not independently add risk points. This keeps the first scoring rule small and explainable.

## Output contract

Human, JSON and Markdown output share the same core service-impact facts.

The report includes:

- `ChangedServices` — directly changed configured services
- `AffectedServices` — direct OpenAPI provider/consumer impacts
- `DownstreamServices` — transitive dependency impacts

Each downstream service carries:

- `Source` — traversal source service
- `Name` — affected downstream service
- `Path` — ordered dependency path from source to affected service
- `Criticality` — configured service criticality
- `Owners` — configured service owners

This preserves the core rule of `impactctl`: every elevated signal should be traceable to explicit evidence.
