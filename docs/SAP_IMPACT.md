# SAP landscape impact — experimental spike

`impactctl sap` tests whether impactctl's deterministic blast-radius model can extend from Git/service changes into an explicitly modeled SAP enterprise landscape.

The spike is intentionally local and credential-free. It does **not** connect directly to SAP CTS/TMS, SAP BTP, Integration Suite or customer systems. Instead, it consumes a versioned YAML manifest containing the changed enterprise component(s) and explicit dependency edges.

The question is deliberately narrow:

> What outside the changed SAP component deserves review, and why?

## Run the synthetic example

```bash
go run ./cmd/impactctl sap --manifest examples/sap-vendor-status.yml
```

Machine-readable output:

```bash
go run ./cmd/impactctl sap --manifest examples/sap-vendor-status.yml --json
```

The example represents an intentionally synthetic chain:

```text
Z_VENDOR_STATUS
  -> BTP_VENDOR_IFLOW
  -> SUPPLIER_PORTAL
  -> VENDOR_APPROVAL
```

No customer or production landscape information is included.

## Manifest schema v1

```yaml
version: 1
change:
  id: DEVK900123
  description: Vendor status API change
  changed:
    - Z_VENDOR_STATUS

nodes:
  - name: Z_VENDOR_STATUS
    kind: sap-object
    criticality: high
    owners: [sap-mm]

  - name: BTP_VENDOR_IFLOW
    kind: integration
    criticality: high
    owners: [integration]
    depends_on: [Z_VENDOR_STATUS]
```

Supported node kinds in the first spike:

- `sap-object`
- `sap-system`
- `integration`
- `api`
- `event`
- `application`
- `data`
- `job`
- `business-process`

Criticality may be omitted or set to `low`, `medium`, `high` or `critical`.

## Dependency semantics

`depends_on` is explicit and directional.

If `BTP_VENDOR_IFLOW` declares:

```yaml
depends_on: [Z_VENDOR_STATUS]
```

then a change to `Z_VENDOR_STATUS` can propagate downstream to `BTP_VENDOR_IFLOW`. Traversal continues transitively through later declared dependencies.

The engine does not infer missing SAP relationships, guess consumers, inspect runtime traffic or fabricate business-process impact.

## Risk score

The spike uses a compact deterministic score so the same manifest always produces the same result:

- base change: `10`
- downstream components: `+8` each, capped at `32`
- affected business processes: `+15` each
- highest declared criticality: `+0 / +10 / +20 / +30`
- four or more distinct component kinds in the affected graph: `+5`
- total capped at `100`

Labels:

- `LOW`: 0–24
- `MEDIUM`: 25–49
- `HIGH`: 50–74
- `CRITICAL`: 75–100

This score is a review-prioritization signal, not an SAP production-approval decision.

## Safety / positioning boundary

This spike does not replace SAP-native change controls, ATC, Clean Core assessment, test-impact tooling, transport governance or human architecture review.

It is testing a different layer: connecting an explicit SAP change to the surrounding integration/application/business-process graph while preserving dependency-path evidence and ownership.

## Follow-on adapters if the spike proves useful

Potential adapters can remain outside the provider-neutral core and translate trusted source evidence into this graph, for example:

1. transport/object metadata from approved SAP export paths;
2. repository-backed Integration Suite artifact metadata;
3. OpenAPI / AsyncAPI contracts already understood by impactctl;
4. explicit enterprise architecture/service catalog relationships;
5. business-process ownership mappings.

Any future adapter should preserve source provenance and fail closed when a relationship cannot be established safely.
