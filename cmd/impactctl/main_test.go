package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/yigitcan-ozturk/impactctl/internal/impact"
	"github.com/yigitcan-ozturk/impactctl/internal/sapimpact"
)

func TestCriticalImpactGoldenFixture(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "impactctl")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build impactctl: %v\n%s", err, out)
	}

	repo := t.TempDir()
	runGit(t, repo, "init", "-b", "main")
	runGit(t, repo, "config", "user.email", "impactctl-test@example.invalid")
	runGit(t, repo, "config", "user.name", "impactctl test")

	writeFixture(t, repo, ".github/CODEOWNERS", "api/ @api-team\ndb/ @data-team\ndeploy/ @platform-team\n")
	writeFixture(t, repo, "README.md", "# fixture\n")
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "baseline")

	runGit(t, repo, "checkout", "-b", "feature/critical-change")
	writeFixture(t, repo, "api/openapi.yaml", "openapi: 3.0.0\ninfo:\n  title: Fixture API\n  version: 1.0.0\n")
	writeFixture(t, repo, "db/migrations/001_add_vendor.sql", "create table vendor(id integer primary key);\n")
	writeFixture(t, repo, "deploy/helm/values.yaml", "replicaCount: 3\n")
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "critical change")

	cmd := exec.Command(bin, "pr", "--base", "main", "--head", "HEAD", "--json")
	cmd.Dir = repo
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("impactctl pr failed: %v\n%s", err, out)
	}

	var report impact.Report
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatalf("decode report: %v\n%s", err, out)
	}

	if report.Risk != "CRITICAL" {
		t.Fatalf("Risk=%s want CRITICAL; report=%+v", report.Risk, report)
	}
	if report.RiskScore != 100 {
		t.Fatalf("RiskScore=%d want 100; report=%+v", report.RiskScore, report)
	}
	wantOwners := []string{"@api-team", "@data-team", "@platform-team"}
	if !reflect.DeepEqual(report.Owners, wantOwners) {
		t.Fatalf("Owners=%v want %v", report.Owners, wantOwners)
	}

	categories := map[string]bool{}
	for _, finding := range report.Findings {
		categories[finding.Category] = true
	}
	for _, want := range []string{"contract", "database", "deployment", "ownership"} {
		if !categories[want] {
			t.Fatalf("missing %q finding in %+v", want, report.Findings)
		}
	}
}

func TestSAPImpactGoldenFixture(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "impactctl")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build impactctl: %v\n%s", err, out)
	}

	manifest := filepath.Join(t.TempDir(), "sap-impact.yml")
	if err := os.WriteFile(manifest, []byte(`
version: 1
change:
  id: DEVK900123
  description: Vendor status API change
  changed: [Z_VENDOR_STATUS]
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
  - name: SUPPLIER_PORTAL
    kind: application
    criticality: medium
    owners: [platform]
    depends_on: [BTP_VENDOR_IFLOW]
  - name: VENDOR_APPROVAL
    kind: business-process
    criticality: critical
    owners: [procurement]
    depends_on: [SUPPLIER_PORTAL]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(bin, "sap", "--manifest", manifest, "--json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("impactctl sap failed: %v\n%s", err, out)
	}

	var report sapimpact.Report
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatalf("decode SAP report: %v\n%s", err, out)
	}
	if report.Risk != "CRITICAL" || report.RiskScore != 84 {
		t.Fatalf("SAP risk=%s %d want CRITICAL 84; report=%+v", report.Risk, report.RiskScore, report)
	}
	if len(report.Downstream) != 3 {
		t.Fatalf("SAP downstream=%d want 3", len(report.Downstream))
	}
	wantProcesses := []string{"VENDOR_APPROVAL"}
	if !reflect.DeepEqual(report.AffectedProcesses, wantProcesses) {
		t.Fatalf("SAP processes=%v want %v", report.AffectedProcesses, wantProcesses)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFixture(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
