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
	_ = fs.Parse(os.Args[2:])

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
	printReport(report)
}

func usage() {
	fmt.Fprintln(os.Stderr, "impactctl — know what your change can break before you merge it")
	fmt.Fprintln(os.Stderr, "\nUsage:\n  impactctl pr [--base main] [--head HEAD] [--json]\n  impactctl version")
}

func printReport(r impact.Report) {
	fmt.Printf("%s IMPACT  (%d/100)\n", r.Risk, r.RiskScore)
	fmt.Println(strings.Repeat("─", 44))
	fmt.Printf("Changed files          %d\n", len(r.Files))
	fmt.Printf("Findings               %d\n", len(r.Findings))
	fmt.Printf("Owner teams            %d\n", len(r.Owners))
	if len(r.Findings) > 0 {
		fmt.Println("\nWhy")
		for _, f := range r.Findings {
			fmt.Printf("! %-12s %s\n", f.Category, f.Detail)
		}
	}
	if len(r.Owners) > 0 {
		fmt.Println("\nSuggested review")
		for _, owner := range r.Owners {
			fmt.Println("→", owner)
		}
	}
}
