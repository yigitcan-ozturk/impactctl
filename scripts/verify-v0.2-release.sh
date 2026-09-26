#!/usr/bin/env bash
# Run from the repository root before proposing a v0.2.0 tag.
set -euo pipefail

version="${1:-v0.2.0-rc}"
if [[ "$version" != v0.2.0* ]]; then
  printf 'Expected a v0.2.0 candidate version, got: %s\n' "$version" >&2
  exit 2
fi

go test ./...
go vet ./...
go build ./cmd/impactctl
bash scripts/build-release.sh "$version"

(cd dist && sha256sum --check checksums.txt)

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
tar -xzf "dist/impactctl_${version}_linux_amd64.tar.gz" -C "$tmp"
actual="$("$tmp/impactctl_${version}_linux_amd64/impactctl" version)"
expected="impactctl $version"
if [[ "$actual" != "$expected" ]]; then
  printf 'Version smoke mismatch: expected %s, got %s\n' "$expected" "$actual" >&2
  exit 1
fi

printf 'PASS: Go tests, vet, build, all release archives, checksums and Linux binary version (%s)\n' "$version"
