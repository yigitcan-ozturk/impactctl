#!/usr/bin/env bash
# Reproducible v0.2 real multi-service validation.
# Usage: IMPACTCTL_REPO=/path/to/impactctl bash scripts/validate-v0.2-external.sh
# Requires: git, go, python3; outbound GitHub access for candidate clone.
set -euo pipefail

ROOT="${IMPACTCTL_REPO:-$(git rev-parse --show-toplevel)}"
OUT="${VALIDATION_OUT:-$ROOT/.validation/v0.2-external}"
CANDIDATE_URL="https://github.com/kaush-utkarsh/microservices-python-demo.git"
mkdir -p "$OUT"
OUT="$(cd "$OUT" && pwd)"
if [[ ! -d "$OUT/candidate/.git" ]]; then
  git clone "$CANDIDATE_URL" "$OUT/candidate"
fi
cd "$OUT/candidate"
git fetch origin main
git checkout --detach a9df3b8b62b0e9e569963989b8ae3c4e1798b150
git rev-parse HEAD > "$OUT/candidate.sha"
(cd "$ROOT" && git rev-parse HEAD) > "$OUT/impactctl.sha"
go version > "$OUT/go-version.txt"
cat > .impactctl.yml <<'YAML'
version: 1
services:
  - name: catalog
    paths: [services/catalog/**]
    criticality: high
  - name: orders
    paths: [services/orders/**]
    criticality: high
    depends_on: [catalog]
  - name: gateway
    paths: [services/gateway/**]
    criticality: medium
    depends_on: [orders, catalog]
YAML
# Real repository source paths, controlled isolated checkout change.
# A controlled source edit is explicitly NOT a historical upstream commit.
git add .impactctl.yml
git -c user.name=impactctl-validation -c user.email=validation@example.invalid commit -qm 'validation: isolated explicit service map'
BASE="$(git rev-parse HEAD)"
printf '\n# impactctl validation: isolated change\n' >> services/catalog/app.py
git add services/catalog/app.py
git -c user.name=impactctl-validation -c user.email=validation@example.invalid commit -qm 'validation: controlled catalog source change'
HEAD="$(git rev-parse HEAD)"
(cd "$ROOT" && go build -o "$OUT/impactctl" ./cmd/impactctl)
"$OUT/impactctl" pr --base "$BASE" --head "$HEAD" > "$OUT/human.txt"
"$OUT/impactctl" pr --base "$BASE" --head "$HEAD" --json > "$OUT/result.json"
"$OUT/impactctl" pr --base "$BASE" --head "$HEAD" --markdown > "$OUT/result.md"
"$OUT/impactctl" pr --base "$BASE" --head "$HEAD" --json > "$OUT/result-rerun.json"
cmp "$OUT/result.json" "$OUT/result-rerun.json"
python3 - "$OUT/result.json" <<'PY'
import json, sys
r=json.load(open(sys.argv[1]))
print("JSON root keys:", sorted(r))
def names(items):
    if not isinstance(items, list):
        raise AssertionError(f"expected list, got {type(items).__name__}")
    return {x if isinstance(x,str) else x.get("Name",x.get("name")) for x in items}
direct=names(r["ChangedServices"])
downstream=names(r["DownstreamServices"])
print("Direct services:", sorted(direct))
print("Downstream services:", sorted(downstream))
assert direct == {"catalog"}, f"unexpected direct services: {direct}"
assert downstream == {"orders","gateway"}, f"unexpected downstream services: {downstream}"
print("PASS: pre-registered catalog impact semantics")
PY
printf 'candidate_base=%s\ncandidate_head=%s\n' "$BASE" "$HEAD" > "$OUT/controlled-change.txt"
echo "Outputs: $OUT"
echo "IMPORTANT: inspect semantic expectations and human/JSON/Markdown parity manually before claiming acceptance."
