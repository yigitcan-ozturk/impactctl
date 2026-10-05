# Online Boutique real-repository validation

## Target

Repository: `GoogleCloudPlatform/microservices-demo` (Online Boutique)

Pinned replay:

- base: `a1377642498628fcb53e9e0f608820a84badff0d`
- head: `25d39dc9a6f84b97b3d84484c9fe392e04e81325`
- selected change: gRPC dependency update in `src/checkoutservice/go.mod` and `src/checkoutservice/go.sum`

## Evidence-backed service model

The validation graph is deliberately minimal. Only relationships supported by repository evidence are declared:

- `frontend` depends on `checkoutservice`: the pinned frontend source configures and opens a gRPC connection through `CHECKOUT_SERVICE_ADDR`.
- `loadgenerator` depends on `frontend`: the target repository documents that the load generator continuously sends realistic shopping-flow requests to frontend.

No other relationship is inferred or declared for this replay.

## Expected impact

Directly changed service:

- `checkoutservice`

Downstream services:

- `checkoutservice → frontend`
- `checkoutservice → frontend → loadgenerator`

No undeclared services should be reported.

## Result

PASS.

The dedicated `Real Repo Validation` workflow builds the current impactctl revision, clones the pinned Online Boutique repository, injects the explicit service map and runs human, JSON and Markdown analysis against the selected commit range.

The replay asserts that all three output formats agree on the core facts:

- exactly two changed files
- changed service is `checkoutservice`
- exactly two downstream services are reported
- downstream paths are deterministic and explainable
- no undeclared service relationship is invented

The first successful validation run also confirmed that normal CI and the repository's own PR Impact workflow remain green.

## Reviewer-scope interpretation

This result demonstrates the intended v0.2 behavior: a seemingly local dependency change in `checkoutservice` expands review awareness to `frontend` and then `loadgenerator` only because those relationships were explicitly evidenced in the configured graph. The tool does not infer additional services from naming or heuristics.

## False-positive / false-negative notes

No material service-relationship false positive or false negative was observed in this replay.
