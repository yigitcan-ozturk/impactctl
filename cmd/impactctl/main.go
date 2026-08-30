package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yigitcan-ozturk/impactctl/internal/impact"
)

var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println("impactctl", version)
		return
	}
	if len(os.Args) < 2 || os.Args[1] != "pr" {
		usage()
		os.Exit(2)
	}

	fs := flag.NewFlagSet("pr", flag.ExitOnError)
	base := fs.String("base", "main", "base git ref")
	head := fs.String("head", "HEAD", "head git ref")
	jsonOut := fs.Bool("json", false, "emit JSON")
	markdownOut := fs.Bool("markdown", false, "emit GitHub-flavored Markdown")
	_ = fs.Parse(os.Args[2:])

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
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(report)
		return
	}
	if *markdownOut {
		fmt.Print(impact.RenderMarkdown(report))
		return
	}
	printReport(report)
}

func usage() {
	fmt.Fprintln(os.Stderr, "impactctl — know what your change can break before you merge it")
	fmt.Fprintln(os.Stderr, "\nUsage:\n  impactctl pr [--base main] [--head HEAD] [--json | --markdown]\n  impactctl version")
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
