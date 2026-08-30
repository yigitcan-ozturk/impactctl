# Service map configuration

`impactctl` v0.2 introduces an optional repository-local service map at `.impactctl.yml`.

The file gives the analyzer explicit evidence for mapping changed repository paths to logical services and for declaring OpenAPI provider/consumer relationships. It is optional: if the file does not exist, v0.1 repository-level analysis continues unchanged.

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
          - mobile

  - name: checkout
    paths:
      - services/checkout/**
    criticality: medium
    owners:
      - "@checkout-team"

  - name: mobile
    paths:
      - apps/mobile/**
    owners:
      - "@mobile-team"
```

In this example, `orders` explicitly provides `api/orders/openapi.yaml`, while `checkout` and `mobile` explicitly consume that contract. A change to the contract reports all three services with their provider/consumer roles.

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
- `openapi` — optional list of OpenAPI contracts provided by this service

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

`impactctl` does not infer API consumers from imports, naming, network calls, repository proximity or other heuristics.

If an OpenAPI file changes and no provider/consumer mapping is declared, the normal v0.1 contract finding still appears, but no service relationship is fabricated.

When a mapping exists, human, JSON and Markdown outputs report:

- provider service
- explicitly declared consumer services
- changed contract path
- configured criticality
- configured service owners

The v0.2 #9 implementation deliberately does not add service relationships to the risk score. Dependency traversal and downstream criticality contribution belong to #11.

## Next boundary

Schema version 1 now supports direct OpenAPI provider/consumer relationships. General service dependency edges and transitive downstream impact remain a separate roadmap step (#11).
