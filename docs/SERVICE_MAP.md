# Service map configuration

`impactctl` v0.2 introduces an optional repository-local service map at `.impactctl.yml`.

The file gives the analyzer explicit evidence for mapping changed repository paths to logical services. It is optional: if the file does not exist, v0.1 repository-level analysis continues unchanged.

## Schema version

The first schema version is `1`.

```yaml
version: 1
services:
  - name: checkout
    paths:
      - services/checkout/**
      - api/checkout/*
    criticality: high
    owners:
      - "@checkout-team"

  - name: inventory
    paths:
      - services/inventory/**
    criticality: medium
    owners:
      - "@inventory-team"
```

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

## Deliberate v0.2 boundary

Schema version 1 does **not** define service dependency edges or API consumer/provider relationships yet. Those are separate roadmap steps (#9 and #11) built on top of this foundation.

Keeping the first schema small makes the configuration contract easier to validate and evolve without weakening the local-first, explicit-evidence design.
