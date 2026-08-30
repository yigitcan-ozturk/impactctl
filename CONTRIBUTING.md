# Contributing to impactctl

Thanks for helping make change impact easier to review.

## Project boundary

`impactctl` keeps a small deterministic core. Contributions should prefer explainable signals over opaque scoring and should not require source code to leave the local or CI environment.

## Development

Requirements:

- Go 1.23+
- Git

Run the verification suite:

```bash
go test ./...
go vet ./...
go build ./cmd/impactctl
```

## Pull requests

Please keep PRs focused and include:

- the problem being solved
- the impact signal or behavior being added/changed
- tests for deterministic behavior
- documentation changes when user-facing behavior changes

Conventional-style commit messages such as `feat:`, `fix:`, `test:`, `docs:` and `chore:` are preferred.

## Good first issues

Issues labeled `good first issue` are intentionally scoped for new contributors. If an issue is unclear, ask in the issue before doing large amounts of work.

## Design principles

- evidence before confidence
- deterministic where possible
- local-first by default
- fail clearly rather than infer silently
- small composable adapters over a monolithic analyzer
