package impact

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestAnalyzeReportsDownstreamPathsAndCriticalRisk(t *testing.T) {
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
    owners: ['@orders-team']
  - name: checkout
    paths: [services/checkout/**]
    criticality: medium
    owners: ['@checkout-team']
    depends_on: [orders]
  - name: reporting
    paths: [services/reporting/**]
    criticality: high
    depends_on: [orders]
  - name: billing
    paths: [services/billing/**]
    criticality: critical
    owners: ['@billing-team']
    depends_on: [checkout]
`
	writeTestFile(t, ".impactctl.yml", config)
	writeTestFile(t, "services/orders/handler.go", "package orders\n\nconst Version = 1\n")
	runGit(t, "add", ".")
	runGit(t, "commit", "-qm", "baseline")
	base := runGit(t, "rev-parse", "HEAD")

	writeTestFile(t, "services/orders/handler.go", "package orders\n\nconst Version = 2\n")
	runGit(t, "add", ".")
	runGit(t, "commit", "-qm", "change orders")

	report, err := Analyze(base, "HEAD")
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if !reflect.DeepEqual(report.ChangedServices, []string{"orders"}) {
		t.Fatalf("ChangedServices = %v, want [orders]", report.ChangedServices)
	}

	var paths []string
	for _, service := range report.DownstreamServices {
		paths = append(paths, strings.Join(service.Path, ">"))
	}
	wantPaths := []string{
		"orders>checkout",
		"orders>reporting",
		"orders>checkout>billing",
	}
	if !reflect.DeepEqual(paths, wantPaths) {
		t.Fatalf("downstream paths = %v, want %v", paths, wantPaths)
	}

	if report.RiskScore != 20 || report.Risk != "MEDIUM" {
		t.Fatalf("risk = %s %d, want MEDIUM 20", report.Risk, report.RiskScore)
	}
	if len(report.Findings) != 1 || report.Findings[0].Category != "downstream" {
		t.Fatalf("Findings = %#v, want one downstream finding", report.Findings)
	}
	if !strings.Contains(report.Findings[0].Detail, "orders -> checkout -> billing") {
		t.Fatalf("downstream finding lacks dependency path: %q", report.Findings[0].Detail)
	}

	billing := report.DownstreamServices[2]
	if billing.Name != "billing" || billing.Criticality != "critical" || !reflect.DeepEqual(billing.Owners, []string{"@billing-team"}) {
		t.Fatalf("billing impact = %#v", billing)
	}
}
