package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yigitcan-ozturk/impactctl/internal/impact"
	"github.com/yigitcan-ozturk/impactctl/internal/sapimpact"
)

var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println("impactctl", version)
		return
	}
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "pr":
		runPR(os.Args[2:])
	case "sap":
		runSAP(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func runPR(args []string) {
	fs := flag.NewFlagSet("pr", flag.ExitOnError)
	base := fs.String("base", "main", "base git ref")
	head := fs.String("head", "HEAD", "head git ref")
	jsonOut := fs.Bool("json", false, "emit JSON")
	markdownOut := fs.Bool("markdown", false, "emit GitHub-flavored Markdown")
	_ = fs.Parse(args)

	if *jsonOut && *markdownOut {
		fmt.Fprintln(os.Stderr, "impactctl: --json and --markdown cannot be used together")
		os.Exit(2)
	}

	report, err := impact.Analyze(*base, *head)
	if err != nil {
		fmt.Fprintln(os.Stderr, "impactctl:", err)
		os.Exit(1)
	}

	if *jsonOut {
		writeJSON(report)
		return
	}
	if *markdownOut {
		fmt.Print(impact.RenderMarkdown(report))
		return
	}
	printReport(report)
}

func runSAP(args []string) {
	fs := flag.NewFlagSet("sap", flag.ExitOnError)
	manifestPath := fs.String("manifest", "", "path to versioned SAP landscape impact YAML")
	jsonOut := fs.Bool("json", false, "emit JSON")
	_ = fs.Parse(args)

	if strings.TrimSpace(*manifestPath) == "" {
		fmt.Fprintln(os.Stderr, "impactctl: sap requires --manifest <file>")
		os.Exit(2)
	}

	manifest, err := sapimpact.Load(*manifestPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "impactctl:", err)
		os.Exit(1)
	}
	report := sapimpact.Analyze(manifest)

	if *jsonOut {
		writeJSON(report)
		return
	}
	printSAPReport(report)
}

func writeJSON(value any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(value)
}

func usage() {
	fmt.Fprintln(os.Stderr, "impactctl — know what your change can break before you merge it")
	fmt.Fprintln(os.Stderr, "\nUsage:\n  impactctl pr [--base main] [--head HEAD] [--json | --markdown]\n  impactctl sap --manifest <file> [--json]\n  impactctl version")
}

func printReport(r impact.Report) {
	fmt.Printf("%s IMPACT  (%d/100)\n", r.Risk, r.RiskScore)
	fmt.Println(strings.Repeat("─", 44))
	fmt.Printf("Changed files          %d\n", len(r.Files))
	fmt.Printf("Findings               %d\n", len(r.Findings))
	fmt.Printf("Owner teams            %d\n", len(r.Owners))
	fmt.Printf("Changed services       %d\n", len(r.ChangedServices))
	fmt.Printf("Contract services      %d\n", len(r.AffectedServices))
	fmt.Printf("Downstream services    %d\n", len(r.DownstreamServices))
	fmt.Printf("AsyncAPI changes       %d\n", len(r.AsyncAPIImpacts))
	if len(r.Findings) > 0 {
		fmt.Println("\nWhy")
		for _, f := range r.Findings {
			fmt.Printf("! %-14s %s\n", f.Category, f.Detail)
		}
	}
	if len(r.AsyncAPIImpacts) > 0 {
		fmt.Println("\nAsyncAPI impact")
		for _, event := range r.AsyncAPIImpacts {
			fmt.Printf("→ %-8s %-8s %-24s %s\n", event.Change, event.Kind, event.Name, event.Detail)
		}
	}
	if len(r.ChangedServices) > 0 {
		fmt.Println("\nChanged services")
		for _, service := range r.ChangedServices {
			fmt.Println("→", service)
		}
	}
	if len(r.AffectedServices) > 0 {
		fmt.Println("\nContract service impact")
		for _, service := range r.AffectedServices {
			fmt.Printf("→ %-8s %-20s via %s", service.Role, service.Name, service.Contract)
			if service.Criticality != "" {
				fmt.Printf(" [%s]", service.Criticality)
			}
			if len(service.Owners) > 0 {
				fmt.Printf(" owners: %s", strings.Join(service.Owners, ", "))
			}
			fmt.Println()
		}
	}
	if len(r.DownstreamServices) > 0 {
		fmt.Println("\nDownstream impact")
		for _, service := range r.DownstreamServices {
			fmt.Printf("→ %-20s via %s", service.Name, strings.Join(service.Path, " -> "))
			if service.Criticality != "" {
				fmt.Printf(" [%s]", service.Criticality)
			}
			if len(service.Owners) > 0 {
				fmt.Printf(" owners: %s", strings.Join(service.Owners, ", "))
			}
			fmt.Println()
		}
	}
	if len(r.Owners) > 0 {
		fmt.Println("\nSuggested review")
		for _, owner := range r.Owners {
			fmt.Println("→", owner)
		}
	}
}

func printSAPReport(r sapimpact.Report) {
	fmt.Printf("%s SAP CHANGE IMPACT  (%d/100)\n", r.Risk, r.RiskScore)
	fmt.Println(strings.Repeat("─", 52))
	fmt.Printf("Change                 %s\n", r.ChangeID)
	if r.Description != "" {
		fmt.Printf("Description            %s\n", r.Description)
	}
	fmt.Printf("Changed components     %d\n", len(r.Changed))
	fmt.Printf("Downstream components  %d\n", len(r.Downstream))
	fmt.Printf("Business processes     %d\n", len(r.AffectedProcesses))
	fmt.Println("Evidence model         explicit dependencies only")

	if len(r.Changed) > 0 {
		fmt.Println("\nChanged SAP / enterprise components")
		for _, node := range r.Changed {
			fmt.Printf("! %-18s %-28s", node.Kind, node.Name)
			if node.Criticality != "" {
				fmt.Printf(" [%s]", node.Criticality)
			}
			if len(node.Owners) > 0 {
				fmt.Printf(" owners: %s", strings.Join(node.Owners, ", "))
			}
			fmt.Println()
		}
	}

	if len(r.Downstream) > 0 {
		fmt.Println("\nEnterprise blast radius")
		for _, item := range r.Downstream {
			fmt.Printf("→ %-18s %-28s via %s", item.Kind, item.Name, strings.Join(item.Path, " -> "))
			if item.Criticality != "" {
				fmt.Printf(" [%s]", item.Criticality)
			}
			if len(item.Owners) > 0 {
				fmt.Printf(" owners: %s", strings.Join(item.Owners, ", "))
			}
			fmt.Println()
		}
	}

	if len(r.AffectedProcesses) > 0 {
		fmt.Println("\nAffected business processes")
		for _, process := range r.AffectedProcesses {
			fmt.Println("→", process)
		}
	}

	if len(r.SuggestedReviewers) > 0 {
		fmt.Println("\nSuggested review")
		for _, owner := range r.SuggestedReviewers {
			fmt.Println("→", owner)
		}
	}
}
