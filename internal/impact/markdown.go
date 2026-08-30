package impact

import (
	"fmt"
	"strings"
)

const commentMarker = "<!-- impactctl:pr-impact -->"

func RenderMarkdown(r Report) string {
	var b strings.Builder
	fmt.Fprintln(&b, commentMarker)
	fmt.Fprintf(&b, "## impactctl — %s IMPACT (%d/100)\n\n", r.Risk, r.RiskScore)
	fmt.Fprintf(&b, "| Signal | Count |\n| --- | ---: |\n| Changed files | %d |\n| Findings | %d |\n| Owner teams | %d |\n| Changed services | %d |\n| Contract services | %d |\n| Downstream services | %d |\n| AsyncAPI changes | %d |\n", len(r.Files), len(r.Findings), len(r.Owners), len(r.ChangedServices), len(r.AffectedServices), len(r.DownstreamServices), len(r.AsyncAPIImpacts))

	if len(r.Findings) > 0 {
		fmt.Fprintln(&b, "\n### Why")
		for _, finding := range r.Findings {
			fmt.Fprintf(&b, "- **%s** — %s\n", finding.Category, finding.Detail)
		}
	}

	if len(r.AsyncAPIImpacts) > 0 {
		fmt.Fprintln(&b, "\n### AsyncAPI impact")
		fmt.Fprintln(&b, "| Change | Kind | Name | Evidence |")
		fmt.Fprintln(&b, "| --- | --- | --- | --- |")
		for _, event := range r.AsyncAPIImpacts {
			fmt.Fprintf(&b, "| **%s** | %s | `%s` | %s |\n", event.Change, event.Kind, event.Name, event.Detail)
		}
	}

	if len(r.ChangedServices) > 0 {
		fmt.Fprintln(&b, "\n### Changed services")
		for _, service := range r.ChangedServices {
			fmt.Fprintf(&b, "- `%s`\n", service)
		}
	}

	if len(r.AffectedServices) > 0 {
		fmt.Fprintln(&b, "\n### Contract service impact")
		fmt.Fprintln(&b, "| Role | Service | Contract | Criticality | Owners |")
		fmt.Fprintln(&b, "| --- | --- | --- | --- | --- |")
		for _, service := range r.AffectedServices {
			criticality := service.Criticality
			if criticality == "" {
				criticality = "—"
			}
			owners := "—"
			if len(service.Owners) > 0 {
				owners = strings.Join(service.Owners, ", ")
			}
			fmt.Fprintf(&b, "| %s | `%s` | `%s` | %s | %s |\n", service.Role, service.Name, service.Contract, criticality, owners)
		}
	}

	if len(r.DownstreamServices) > 0 {
		fmt.Fprintln(&b, "\n### Downstream impact")
		fmt.Fprintln(&b, "| Service | Dependency path | Criticality | Owners |")
		fmt.Fprintln(&b, "| --- | --- | --- | --- |")
		for _, service := range r.DownstreamServices {
			criticality := service.Criticality
			if criticality == "" {
				criticality = "—"
			}
			owners := "—"
			if len(service.Owners) > 0 {
				owners = strings.Join(service.Owners, ", ")
			}
			fmt.Fprintf(&b, "| `%s` | `%s` | %s | %s |\n", service.Name, strings.Join(service.Path, " → "), criticality, owners)
		}
	}

	if len(r.Owners) > 0 {
		fmt.Fprintln(&b, "\n### Suggested review")
		for _, owner := range r.Owners {
			fmt.Fprintf(&b, "- `%s`\n", owner)
		}
	}

	fmt.Fprintln(&b, "\n_Deterministic local analysis. No source code is sent to an external service._")
	return strings.TrimSpace(b.String()) + "\n"
}

func CommentMarker() string {
	return commentMarker
}
