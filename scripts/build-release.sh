#!/usr/bin/env bash
set -euo pipefail

version="${1:?usage: build-release.sh VERSION [goos/goarch ...]}"
shift || true

if [ "$#" -gt 0 ]; then
  targets=("$@")
else
  targets=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
  )
fi

rm -rf dist .release-work
mkdir -p dist .release-work

for target in "${targets[@]}"; do
  goos="${target%/*}"
  goarch="${target#*/}"

  if [ -z "$goos" ] || [ -z "$goarch" ] || [ "$goos" = "$target" ]; then
    echo "invalid target: $target (expected goos/goarch)" >&2
    exit 2
  fi

  ext=""
  if [ "$goos" = "windows" ]; then
    ext=".exe"
  fi

  name="impactctl_${version}_${goos}_${goarch}"
  package_dir=".release-work/${name}"
  mkdir -p "$package_dir"

  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
    go build -trimpath \
    -ldflags "-s -w -X main.version=${version}" \
    -o "${package_dir}/impactctl${ext}" ./cmd/impactctl

  cp LICENSE README.md "$package_dir/"

  if [ "$goos" = "windows" ]; then
    (cd .release-work && zip -qr "../dist/${name}.zip" "$name")
  else
    tar -C .release-work -czf "dist/${name}.tar.gz" "$name"
  fi
done

(
  cd dist
  sha256sum -- * > checksums.txt
)

echo "release artifacts written to dist/"
