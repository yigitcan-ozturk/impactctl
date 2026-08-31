# SAP landscape impact — practitioner validation pack

This pack is for architecture review, not product promotion.

## Goal

Validate whether `impactctl sap` identifies a useful and explainable review scope for SAP-originated changes that propagate into integrations, applications and business processes.

## Scenarios

1. `examples/sap-material-api.yml`
   - SAP material API -> Integration Suite -> product data hub -> commerce catalog -> material publishing
2. `examples/sap-idoc-warehouse.yml`
   - SAP delivery IDoc mapping -> integration flow -> warehouse execution -> shipping job -> outbound fulfilment
3. `examples/sap-event-maintenance.yml`
   - SAP equipment event -> event bridge -> API -> field service application -> breakdown response

All scenarios are synthetic and contain no customer or production identifiers.

## Run

```bash
go run ./cmd/impactctl sap --manifest examples/sap-material-api.yml
go run ./cmd/impactctl sap --manifest examples/sap-idoc-warehouse.yml
go run ./cmd/impactctl sap --manifest examples/sap-event-maintenance.yml
```

Machine-readable output can be generated with `--json`.

## Reviewer questions

For each scenario, review only the architecture logic:

1. Is the downstream review scope directionally correct?
2. Is any important dependency class missing?
3. Is any path too broad or misleading?
4. Are owner/reviewer hints useful in a real release decision?
5. Would this output change who you involve before production?
6. Which source of truth would you trust for generating these edges in a real landscape?
7. Which first adapter would create the most practical value: transport/object metadata, Integration Suite/Git metadata, API/event contracts, or EA/service-catalog relationships?

## What a valid review is not

- A star or like is not validation.
- Agreement with the demo is not enough; reviewers should identify missing/incorrect assumptions.
- The score is a review-prioritization signal, not production approval.
- The engine must not infer dependencies that are not supported by evidence.

## Evidence record

For each practitioner review record privately or in a sanitized issue comment:

- reviewer role/domain;
- scenario reviewed;
- expected blast radius before seeing output;
- missing/excess paths;
- ownership usefulness;
- preferred first production adapter;
- whether the output would change review/release behavior.

Use the collected evidence for the decision gate in issue #28. Do not commit confidential customer or system information.
