package impact

import (
	"os"
	"reflect"
	"testing"
)

func TestAnalyzeAsyncAPIAdditiveChange(t *testing.T) {
	root, base := initAsyncAPIFixture(t, asyncAPIBase())
	withWorkingDir(t, root, func() {
		writeTestFile(t, "contracts/asyncapi.yaml", `asyncapi: 3.0.0
info:
  title: Events
  version: 1.1.0
channels:
  order.created:
    address: order.created
  order.cancelled:
    address: order.cancelled
components:
  messages:
    OrderCreated:
      payload:
        $ref: '#/components/schemas/Order'
    OrderCancelled:
      payload:
        $ref: '#/components/schemas/Cancellation'
  schemas:
    Order:
      type: object
    Cancellation:
      type: object
`)
		runGit(t, "add", ".")
		runGit(t, "commit", "-qm", "add event contract entities")

		report, err := Analyze(base, "HEAD")
		if err != nil {
			t.Fatalf("Analyze() error = %v", err)
		}
		if report.RiskScore != 5 || report.Risk != "LOW" {
			t.Fatalf("risk = %s %d, want LOW 5", report.Risk, report.RiskScore)
		}
		var got []string
		for _, impact := range report.AsyncAPIImpacts {
			got = append(got, impact.Change+":"+impact.Kind+":"+impact.Name)
		}
		want := []string{
			"ADDITIVE:channel:order.cancelled",
			"ADDITIVE:message:OrderCancelled",
			"ADDITIVE:schema:Cancellation",
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("AsyncAPIImpacts = %v, want %v", got, want)
		}
		if len(report.Findings) != 1 || report.Findings[0].Category != "event-additive" {
			t.Fatalf("Findings = %#v, want event-additive", report.Findings)
		}
	})
}

func TestAnalyzeAsyncAPIBreakingRemoval(t *testing.T) {
	initial := `asyncapi: 3.0.0
info:
  title: Events
  version: 1.0.0
channels:
  order.created:
    address: order.created
  order.cancelled:
    address: order.cancelled
components:
  messages:
    OrderCreated:
      payload:
        $ref: '#/components/schemas/Order'
    OrderCancelled:
      payload:
        $ref: '#/components/schemas/Cancellation'
  schemas:
    Order:
      type: object
    Cancellation:
      type: object
`
	root, base := initAsyncAPIFixture(t, initial)
	withWorkingDir(t, root, func() {
		writeTestFile(t, "contracts/asyncapi.yaml", asyncAPIBase())
		runGit(t, "add", ".")
		runGit(t, "commit", "-qm", "remove event contract entities")

		report, err := Analyze(base, "HEAD")
		if err != nil {
			t.Fatalf("Analyze() error = %v", err)
		}
		if report.RiskScore != 35 || report.Risk != "MEDIUM" {
			t.Fatalf("risk = %s %d, want MEDIUM 35", report.Risk, report.RiskScore)
		}
		var got []string
		for _, impact := range report.AsyncAPIImpacts {
			got = append(got, impact.Change+":"+impact.Kind+":"+impact.Name)
		}
		want := []string{
			"BREAKING:channel:order.cancelled",
			"BREAKING:message:OrderCancelled",
			"BREAKING:schema:Cancellation",
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("AsyncAPIImpacts = %v, want %v", got, want)
		}
		if len(report.Findings) != 1 || report.Findings[0].Category != "event-breaking" {
			t.Fatalf("Findings = %#v, want event-breaking", report.Findings)
		}
	})
}

func TestAnalyzeAsyncAPIChangedEntityRequiresReview(t *testing.T) {
	root, base := initAsyncAPIFixture(t, asyncAPIBase())
	withWorkingDir(t, root, func() {
		writeTestFile(t, "contracts/asyncapi.yaml", `asyncapi: 3.0.0
info:
  title: Events
  version: 1.0.0
channels:
  order.created:
    address: order.created
components:
  messages:
    OrderCreated:
      payload:
        $ref: '#/components/schemas/Order'
  schemas:
    Order:
      type: object
      required: [id]
      properties:
        id:
          type: string
`)
		runGit(t, "add", ".")
		runGit(t, "commit", "-qm", "change event schema")

		report, err := Analyze(base, "HEAD")
		if err != nil {
			t.Fatalf("Analyze() error = %v", err)
		}
		if report.RiskScore != 15 || report.Risk != "LOW" {
			t.Fatalf("risk = %s %d, want LOW 15", report.Risk, report.RiskScore)
		}
		if len(report.AsyncAPIImpacts) != 1 || report.AsyncAPIImpacts[0].Change != AsyncAPIReview || report.AsyncAPIImpacts[0].Kind != "schema" {
			t.Fatalf("AsyncAPIImpacts = %#v, want one REVIEW schema", report.AsyncAPIImpacts)
		}
		if len(report.Findings) != 1 || report.Findings[0].Category != "event-review" {
			t.Fatalf("Findings = %#v, want event-review", report.Findings)
		}
	})
}

func TestAsyncAPIDetectorIgnoresSourceFilenames(t *testing.T) {
	for _, path := range []string{"internal/impact/asyncapi.go", "asyncapi_test.go", "docs/asyncapi.md"} {
		if isAsyncAPIPath(path) {
			t.Fatalf("isAsyncAPIPath(%q)=true, want false", path)
		}
	}
	for _, path := range []string{"asyncapi.yaml", "contracts/events.asyncapi.yml", "spec/asyncapi.json"} {
		if !isAsyncAPIPath(path) {
			t.Fatalf("isAsyncAPIPath(%q)=false, want true", path)
		}
	}
}

func initAsyncAPIFixture(t *testing.T, document string) (string, string) {
	t.Helper()
	root := t.TempDir()
	var base string
	withWorkingDir(t, root, func() {
		runGit(t, "init", "-q")
		runGit(t, "config", "user.email", "impactctl-test@example.invalid")
		runGit(t, "config", "user.name", "impactctl test")
		writeTestFile(t, "contracts/asyncapi.yaml", document)
		runGit(t, "add", ".")
		runGit(t, "commit", "-qm", "baseline")
		base = runGit(t, "rev-parse", "HEAD")
	})
	return root, base
}

func withWorkingDir(t *testing.T, root string, fn func()) {
	t.Helper()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(oldWD); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	}()
	fn()
}

func asyncAPIBase() string {
	return `asyncapi: 3.0.0
info:
  title: Events
  version: 1.0.0
channels:
  order.created:
    address: order.created
components:
  messages:
    OrderCreated:
      payload:
        $ref: '#/components/schemas/Order'
  schemas:
    Order:
      type: object
`
}
