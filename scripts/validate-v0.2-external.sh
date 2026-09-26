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
printf '
# impactctl validation: isolated change
' >> services/catalog/app.py
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
    if items is None:
        items = []
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
# Assert human/JSON/Markdown parity on real upstream catalog scenario.
python3 - "$OUT/result.json" "$OUT/human.txt" "$OUT/result.md" <<'PY'
import json,sys
r=json.load(open(sys.argv[1]))
human=open(sys.argv[2]).read()
md=open(sys.argv[3]).read()
changed=r["ChangedServices"] or []
downstream=r["DownstreamServices"] or []
assert "Changed services       1" in human
assert "Downstream services    2" in human
assert "| Changed services | 1 |" in md
assert "| Downstream services | 2 |" in md
assert "Changed services" in human and "### Changed services" in md
assert "Downstream impact" in human and "### Downstream impact" in md
for name in changed:
    assert "→ "+name in human, ("human changed",name)
    assert "`"+name+"`" in md, ("markdown changed",name)
for item in downstream:
    name=item["Name"]
    path=item["Path"]
    assert name in human and " -> ".join(path) in human, ("human downstream",item)
    assert "`"+name+"`" in md and " → ".join(path) in md, ("markdown downstream",item)
assert str(r["RiskScore"])+"/100" in human and str(r["RiskScore"])+"/100" in md
print("PASS: human/JSON/Markdown core service facts, paths and risk score")
PY
printf 'candidate_base=%s
candidate_head=%s
' "$BASE" "$HEAD" > "$OUT/controlled-change.txt"
echo "Outputs: $OUT"

# Verify the opposite direction and a terminal gateway node in the same
# pinned REAL repository, with independent controlled commits.
for SERVICE in orders gateway; do
  git reset --hard "$BASE" >/dev/null
  printf '
# impactctl validation: %s-only controlled change
' "$SERVICE" >> "services/$SERVICE/app.py"
  git add "services/$SERVICE/app.py"
  git -c user.name=impactctl-validation -c user.email=validation@example.invalid commit -qm "validation: controlled $SERVICE source change"
  CASE_HEAD="$(git rev-parse HEAD)"
  "$OUT/impactctl" pr --base "$BASE" --head "$CASE_HEAD" --json > "$OUT/$SERVICE.json"
  "$OUT/impactctl" pr --base "$BASE" --head "$CASE_HEAD" --json > "$OUT/$SERVICE-rerun.json"
  cmp "$OUT/$SERVICE.json" "$OUT/$SERVICE-rerun.json"
  python3 - "$OUT/$SERVICE.json" "$SERVICE" <<'PY'
import json,sys
r=json.load(open(sys.argv[1]))
service=sys.argv[2]
def names(items):
    if items is None: items=[]
    assert isinstance(items,list), type(items)
    return {x if isinstance(x,str) else x.get("Name",x.get("name")) for x in items}
expected={"orders":{"gateway"},"gateway":set()}
direct=names(r["ChangedServices"])
downstream=names(r["DownstreamServices"])
assert direct=={service}, (service,direct)
assert downstream==expected[service], (service,downstream)
print("PASS:",service,"direct",sorted(direct),"downstream",sorted(downstream))
PY
done

# Legacy no-config regression: same pinned repository without a service map.
# Do not silently claim byte-for-byte v0.1 parity; require explicit no-service
# behavior, and keep the original v0.1 regression suite as a separate gate.
git reset --hard a9df3b8b62b0e9e569963989b8ae3c4e1798b150 >/dev/null
printf '
# impactctl validation: legacy no-config change
' >> services/catalog/app.py
git add services/catalog/app.py
git -c user.name=impactctl-validation -c user.email=validation@example.invalid commit -qm 'validation: legacy no-config catalog change'
LEGACY_HEAD="$(git rev-parse HEAD)"
"$OUT/impactctl" pr --base a9df3b8b62b0e9e569963989b8ae3c4e1798b150 --head "$LEGACY_HEAD" --json > "$OUT/no-config.json"
python3 - "$OUT/no-config.json" <<'PY'
import json,sys
r=json.load(open(sys.argv[1]))
assert not r.get("ChangedServices"), r.get("ChangedServices")
assert not r.get("DownstreamServices"), r.get("DownstreamServices")
assert r.get("Files"), "legacy diff should still be detected"
print("PASS: no-config repository analysis preserved, no invented service graph")
PY
echo "PASS: catalog/orders/gateway/no-config cases; inspect artifacts for output parity."

