package sapimpact

import (
	"reflect"
	"testing"
)

func TestAnalyzeTraversesSAPLandscapeToBusinessProcess(t *testing.T) {
	manifest, err := Parse([]byte(`
version: 1
change:
  id: DEVK900123
  description: Vendor status API change
  changed:
    - Z_VENDOR_STATUS
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
`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	report := Analyze(manifest)
	if report.Risk != "CRITICAL" || report.RiskScore != 84 {
		t.Fatalf("risk = %s %d, want CRITICAL 84", report.Risk, report.RiskScore)
	}
	if len(report.Downstream) != 3 {
		t.Fatalf("downstream count = %d, want 3", len(report.Downstream))
	}

	gotPath := report.Downstream[2].Path
	wantPath := []string{"Z_VENDOR_STATUS", "BTP_VENDOR_IFLOW", "SUPPLIER_PORTAL", "VENDOR_APPROVAL"}
	if !reflect.DeepEqual(gotPath, wantPath) {
		t.Fatalf("path = %#v, want %#v", gotPath, wantPath)
	}

	wantProcesses := []string{"VENDOR_APPROVAL"}
	if !reflect.DeepEqual(report.AffectedProcesses, wantProcesses) {
		t.Fatalf("processes = %#v, want %#v", report.AffectedProcesses, wantProcesses)
	}

	wantOwners := []string{"integration", "platform", "procurement", "sap-mm"}
	if !reflect.DeepEqual(report.SuggestedReviewers, wantOwners) {
		t.Fatalf("reviewers = %#v, want %#v", report.SuggestedReviewers, wantOwners)
	}
}

func TestParseRejectsUnknownDependency(t *testing.T) {
	_, err := Parse([]byte(`
version: 1
change:
  id: DEVK900124
  changed: [Z_API]
nodes:
  - name: Z_API
    kind: sap-object
  - name: IFlow
    kind: integration
    depends_on: [MISSING_NODE]
`))
	if err == nil {
		t.Fatal("Parse() expected error for unknown dependency")
	}
}

func TestParseRejectsUnknownFields(t *testing.T) {
	_, err := Parse([]byte(`
version: 1
change:
  id: DEVK900125
  changed: [Z_API]
nodes:
  - name: Z_API
    kind: sap-object
    guessed_dependency: something
`))
	if err == nil {
		t.Fatal("Parse() expected error for unknown field")
	}
}
