package impact

import (
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	report := Report{
		Risk:      "HIGH",
		RiskScore: 55,
		Files:     []string{"api/openapi.yaml", "db/migrations/001.sql"},
		Findings: []Finding{
			{Category: "contract", Detail: "api/openapi.yaml changes an API contract", Weight: 30},
			{Category: "database", Detail: "db/migrations/001.sql looks like a database migration", Weight: 35},
		},
		Owners: []string{"@api-team", "@data-team"},
	}

	got := RenderMarkdown(report)
	for _, want := range []string{
		CommentMarker(),
		"## impactctl — HIGH IMPACT (55/100)",
		"| Changed files | 2 |",
		"**contract** — api/openapi.yaml changes an API contract",
		"`@api-team`",
		"No source code is sent to an external service.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderMarkdown output missing %q:\n%s", want, got)
		}
	}
}

func TestRenderMarkdownAlwaysEndsWithNewline(t *testing.T) {
	if got := RenderMarkdown(Report{Risk: "LOW"}); !strings.HasSuffix(got, "\n") {
		t.Fatalf("RenderMarkdown should end with newline: %q", got)
	}
}
