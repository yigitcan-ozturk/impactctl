# impactctl traction

This document tracks evidence that `impactctl` is being discovered, tried and validated. Traction is treated as product evidence, not as a vanity gate for engineering work.

## Baseline — 2026-08-30 20:54 TRT

The baseline was captured shortly after the first public release.

| Signal | Baseline |
|---|---:|
| GitHub stars | 0 |
| Forks | 0 |
| Public watchers | 0 |
| Subscribers | 0 |
| Open issues | 5 |
| v0.1.0 release asset downloads | 0 |
| External repository validations recorded | 0 |
| External contributor PRs recorded | 0 |

Repository creation: 2026-08-30. `v0.1.0` was published on 2026-08-30 at 15:59:40 UTC.

The initial release contains cross-platform artifacts for Linux, macOS and Windows plus checksums. At baseline, each published release asset reports zero downloads.

## What we measure

### Discovery

- stars
- public watchers/subscribers
- forks
- external mentions that can be tied back to the project

### Trial / adoption

- release asset downloads
- unique clones / cloners when GitHub traffic data is available
- external repositories where `impactctl` is actually run
- GitHub Action installations or copied workflow usage when verifiable

### Engagement

- issues opened by people other than the repository owner
- pull requests from external contributors
- issue/PR comments and reactions from external users
- repeat contributors or repeat testers

### Product validation

The strongest signal is not a star. It is a real repository producing a useful, explainable finding that changes review behavior.

For each real-world validation, record:

1. repository type / stack
2. command or workflow used
3. detected impact
4. whether the finding was correct
5. false-positive / false-negative notes
6. whether the output changed a reviewer decision or review scope
7. follow-up feature request

## Near-term traction checkpoints

### Checkpoint A — first external proof

- first star from a non-owner account
- first release download
- first external repository validation

### Checkpoint B — repeatability

- at least 3 independently selected repositories tested
- at least 2 different stack patterns represented
- recurring finding categories are understood
- false-positive notes are converted into scoped issues where appropriate

### Checkpoint C — early community signal

- first external issue or pull request
- first fork used for real experimentation
- evidence that at least one user returns after an initial trial

## v0.2 relationship

Traction tracking runs in parallel with v0.2 engineering. v0.2 should not chase stars; it should improve the probability that a real trial becomes a repeatable workflow.

The v0.2 service-impact release should therefore be evaluated on two axes:

- **engineering gate:** deterministic service-impact model and release quality
- **validation gate:** real repositories demonstrate that provider/consumer and downstream-impact evidence is useful

See [`V0.2_EXECUTION.md`](V0.2_EXECUTION.md) for the engineering sequence and release gate.

## Snapshot rule

When a new snapshot is taken, append a dated section rather than rewriting the baseline. Record both the absolute values and the delta from the previous snapshot.
