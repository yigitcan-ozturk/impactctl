package impact

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAnalyzeReportsExplicitOpenAPIProviderConsumers(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	runGit(t, "init", "-q")
	runGit(t, "config", "user.email", "impactctl-test@example.invalid")
	runGit(t, "config", "user.name", "impactctl test")

	config := `version: 1
services:
  - name: orders
    paths: [services/orders/**]
    criticality: high
    owners: ['@orders-team']
    openapi:
      - path: api/orders/openapi.yaml
        consumers: [checkout, mobile]
  - name: checkout
    paths: [services/checkout/**]
    criticality: medium
    owners: ['@checkout-team']
  - name: mobile
    paths: [apps/mobile/**]
    owners: ['@mobile-team']
`
	writeTestFile(t, ".impactctl.yml", config)
	writeTestFile(t, "api/orders/openapi.yaml", "openapi: 3.0.0\ninfo:\n  title: Orders\n  version: 1.0.0\n")
	runGit(t, "add", ".")
	runGit(t, "commit", "-qm", "baseline")
	base := runGit(t, "rev-parse", "HEAD")

	writeTestFile(t, "api/orders/openapi.yaml", "openapi: 3.0.0\ninfo:\n  title: Orders\n  version: 1.1.0\n")
	runGit(t, "add", ".")
	runGit(t, "commit", "-qm", "change contract")

	report, err := Analyze(base, "HEAD")
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if report.RiskScore != 30 || report.Risk != "MEDIUM" {
		t.Fatalf("risk = %s %d, want MEDIUM 30", report.Risk, report.RiskScore)
	}

	var got []string
	for _, service := range report.AffectedServices {
		got = append(got, service.Role+":"+service.Name+":"+service.Contract)
	}
	want := []string{
		"provider:orders:api/orders/openapi.yaml",
		"consumer:checkout:api/orders/openapi.yaml",
		"consumer:mobile:api/orders/openapi.yaml",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AffectedServices = %v, want %v", got, want)
	}
	if gotOwners := report.AffectedServices[0].Owners; !reflect.DeepEqual(gotOwners, []string{"@orders-team"}) {
		t.Fatalf("provider owners = %v", gotOwners)
	}
}

func runGit(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return stringTrimSpace(string(out))
}

func writeTestFile(t *testing.T, name, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func stringTrimSpace(value string) string {
	for len(value) > 0 {
		last := value[len(value)-1]
		if last != '\n' && last != '\r' && last != ' ' && last != '\t' {
			break
		}
		value = value[:len(value)-1]
	}
	return value
}
