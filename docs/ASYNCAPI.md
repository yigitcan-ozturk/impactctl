# AsyncAPI impact analysis

`impactctl` v0.2 can compare common AsyncAPI YAML/JSON contract files across the pull-request base and head revisions.

The implementation is intentionally conservative. It does not attempt to become a complete AsyncAPI semantic compatibility engine. It extracts a small set of named contract entities and reports only what can be determined locally and explainably.

## Detection

Native AsyncAPI analysis applies to YAML, YML or JSON files whose filename contains `asyncapi`, for example:

- `asyncapi.yaml`
- `events.asyncapi.yml`
- `contracts/asyncapi.json`

Source files such as `asyncapi.go` and documentation such as `asyncapi.md` are not treated as event contracts.

A detected document must contain the root `asyncapi` version field on at least one side of the comparison.

## Compared entities

`impactctl` compares named entities in:

- `channels`
- `components.messages`
- `components.schemas`

The base and head files are read directly from Git using the supplied refs. No external service or network call is required.

## Classification

Each changed entity receives one stable classification:

### `ADDITIVE`

A named channel, message or schema exists in the head revision but not the base revision.

This produces an `event-additive` finding for the contract file. The initial risk contribution is +5.

### `BREAKING`

A named channel, message or schema exists in the base revision but not the head revision.

A removal may also represent a rename because a rename appears as one removed name and one added name. `impactctl` therefore treats the removed side conservatively as potentially breaking.

This produces an `event-breaking` finding for the contract file with +35 risk.

### `REVIEW`

A named channel, message or schema exists in both revisions but its parsed content changed.

`impactctl` does not guess whether the semantic change is compatible. It reports the entity for review instead.

Changes elsewhere in an AsyncAPI document that cannot be safely reduced to the supported entity sets also produce `REVIEW` evidence.

This produces an `event-review` finding with +15 risk.

## Output contract

The report includes `AsyncAPIImpacts`. Each entry contains:

- `Path` — changed AsyncAPI document
- `Kind` — `channel`, `message`, `schema`, or `contract`
- `Name` — named entity or contract path
- `Change` — `ADDITIVE`, `BREAKING`, or `REVIEW`
- `Detail` — human-readable evidence

Human, JSON and Markdown output use the same impact records.

## File-level scoring

Risk is applied once per AsyncAPI document, using the most conservative classification found in that document:

- any `BREAKING` entity → `event-breaking` / +35
- otherwise any `REVIEW` entity → `event-review` / +15
- otherwise additive-only changes → `event-additive` / +5

Individual entities remain visible as evidence without multiplying the score for every removed schema or channel.

## Limitations

The initial v0.2 analyzer does not claim to determine compatibility for all AsyncAPI protocol bindings, message traits, schema dialects, operation semantics or broker-specific behavior.

A modified entity is therefore deliberately `REVIEW` unless the change is a clear name-level addition or removal. Deep domain semantics can be added later through focused adapters without weakening the deterministic core.
