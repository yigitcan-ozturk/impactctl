#!/usr/bin/env bash
set -euo pipefail

TARGET_REPO="https://github.com/GoogleCloudPlatform/microservices-demo.git"
BASE_SHA="a1377642498628fcb53e9e0f608820a84badff0d"
HEAD_SHA="25d39dc9a6f84b97b3d84484c9fe392e04e81325"
IMPACTCTL_BIN="${IMPACTCTL_BIN:-$(pwd)/impactctl}"

if [[ ! -x "$IMPACTCTL_BIN" ]]; then
  echo "impactctl binary not found or not executable: $IMPACTCTL_BIN" >&2
  exit 1
fi

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

target="$workdir/online-boutique"
echo "Cloning pinned Online Boutique validation target..."
git clone --quiet "$TARGET_REPO" "$target"
git -C "$target" checkout --quiet "$HEAD_SHA"

if ! git -C "$target" cat-file -e "${BASE_SHA}^{commit}"; then
  echo "Pinned base commit is not available in target clone" >&2
  exit 1
fi

# Evidence boundary for this validation graph:
# - frontend directly configures and opens a gRPC connection to CHECKOUT_SERVICE_ADDR
#   in src/frontend/main.go at the pinned revision.
# - the target repository README states that loadgenerator continuously sends
#   realistic shopping-flow requests to frontend.
# No other service relationship is declared for this replay.
cat > "$target/.impactctl.yml" <<'YAML'
version: 1
services:
  - name: checkoutservice
    paths:
      - src/checkoutservice/**

  - name: frontend
    paths:
      - src/frontend/**
    depends_on:
      - checkoutservice

  - name: loadgenerator
    paths:
      - src/loadgenerator/**
    depends_on:
      - frontend
YAML

json_out="$workdir/report.json"
human_out="$workdir/report.txt"
markdown_out="$workdir/report.md"

(
  cd "$target"
  "$IMPACTCTL_BIN" pr --base "$BASE_SHA" --head "$HEAD_SHA" --json > "$json_out"
  "$IMPACTCTL_BIN" pr --base "$BASE_SHA" --head "$HEAD_SHA" > "$human_out"
  "$IMPACTCTL_BIN" pr --base "$BASE_SHA" --head "$HEAD_SHA" --markdown > "$markdown_out"
)

python3 - "$json_out" <<'PY'
import json
import sys

with open(sys.argv[1], encoding="utf-8") as fh:
    report = json.load(fh)

assert report["Files"] == [
    "src/checkoutservice/go.mod",
    "src/checkoutservice/go.sum",
], report["Files"]
assert report["ChangedServices"] == ["checkoutservice"], report["ChangedServices"]
assert report["Risk"] == "LOW", report["Risk"]
assert report["RiskScore"] == 0, report["RiskScore"]

actual = [(item["Name"], item["Path"]) for item in report["DownstreamServices"]]
expected = [
    ("frontend", ["checkoutservice", "frontend"]),
    ("loadgenerator", ["checkoutservice", "frontend", "loadgenerator"]),
]
assert actual == expected, actual
assert report["AffectedServices"] is None or report["AffectedServices"] == [], report["AffectedServices"]
PY

grep -Fq "LOW IMPACT  (0/100)" "$human_out"
grep -Fq "Changed services       1" "$human_out"
grep -Fq "Downstream services    2" "$human_out"
grep -Fq "checkoutservice -> frontend" "$human_out"
grep -Fq "checkoutservice -> frontend -> loadgenerator" "$human_out"

grep -Fq "## impactctl — LOW IMPACT (0/100)" "$markdown_out"
grep -Fq "| Changed services | 1 |" "$markdown_out"
grep -Fq "| Downstream services | 2 |" "$markdown_out"
grep -Fq 'checkoutservice → frontend' "$markdown_out"
grep -Fq 'checkoutservice → frontend → loadgenerator' "$markdown_out"

echo
cat "$human_out"
echo
printf '%s\n' "--- validation summary ---"
printf 'target: GoogleCloudPlatform/microservices-demo\n'
printf 'base:   %s\n' "$BASE_SHA"
printf 'head:   %s\n' "$HEAD_SHA"
printf '%s\n' 'result: PASS — checkoutservice change expands to frontend and loadgenerator with explicit dependency evidence'
