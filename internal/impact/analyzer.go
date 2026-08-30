package impact

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yigitcan-ozturk/impactctl/internal/servicemap"
)

type Finding struct {
	Category string
	Detail   string
	Weight   int
}

type ServiceImpact struct {
	Name        string
	Role        string
	Contract    string
	Criticality string
	Owners      []string
}

type Report struct {
	Base             string
	Head             string
	Files            []string
	Findings         []Finding
	Owners           []string
	AffectedServices []ServiceImpact
	RiskScore        int
	Risk             string
}

func Analyze(base, head string) (Report, error) {
	files, err := changedFiles(base, head)
	if err != nil {
		return Report{}, err
	}

	serviceMap, hasServiceMap, err := servicemap.Load(".")
	if err != nil {
		return Report{}, err
	}

	report := Report{Base: base, Head: head, Files: files}
	owners := map[string]struct{}{}
	serviceImpacts := map[string]ServiceImpact{}

	for _, f := range files {
		lower := strings.ToLower(f)
		var openAPIImpacts []servicemap.OpenAPIImpact
		if hasServiceMap {
			openAPIImpacts = serviceMap.OpenAPIImpactsForPath(f)
		}

		switch {
		case isOpenAPI(lower) || len(openAPIImpacts) > 0:
			report.Findings = append(report.Findings, Finding{"contract", f + " changes an API contract", 30})
		case isMigration(lower):
			report.Findings = append(report.Findings, Finding{"database", f + " looks like a database migration", 35})
		case isDeployment(lower):
			report.Findings = append(report.Findings, Finding{"deployment", f + " changes deployment/infrastructure", 25})
		case isWorkflow(lower):
			report.Findings = append(report.Findings, Finding{"ci", f + " changes CI/CD behavior", 15})
		case isConfig(lower):
			report.Findings = append(report.Findings, Finding{"config", f + " changes runtime/configuration behavior", 15})
		}

		for _, relationship := range openAPIImpacts {
			provider := newServiceImpact(relationship.Provider, "provider", f)
			serviceImpacts[serviceImpactKey(provider)] = provider
			for _, consumerService := range relationship.Consumers {
				consumer := newServiceImpact(consumerService, "consumer", f)
				serviceImpacts[serviceImpactKey(consumer)] = consumer
			}
		}
	}

	if len(files) >= 20 {
		report.Findings = append(report.Findings, Finding{"scope", fmt.Sprintf("%d files changed", len(files)), 20})
	} else if len(files) >= 8 {
		report.Findings = append(report.Findings, Finding{"scope", fmt.Sprintf("%d files changed", len(files)), 10})
	}

	if codeownersPath, ok := discoverCodeowners("."); ok {
		if parsed, err := loadCodeowners(codeownersPath); err == nil {
			for _, f := range files {
				for _, owner := range parsed.match(f) {
					owners[owner] = struct{}{}
				}
			}
		}
	}
	for owner := range owners {
		report.Owners = append(report.Owners, owner)
	}
	sort.Strings(report.Owners)

	for _, serviceImpact := range serviceImpacts {
		report.AffectedServices = append(report.AffectedServices, serviceImpact)
	}
	sort.Slice(report.AffectedServices, func(i, j int) bool {
		left, right := report.AffectedServices[i], report.AffectedServices[j]
		if left.Contract != right.Contract {
			return left.Contract < right.Contract
		}
		if left.Role != right.Role {
			return serviceRoleRank(left.Role) < serviceRoleRank(right.Role)
		}
		return left.Name < right.Name
	})

	for _, finding := range report.Findings {
		report.RiskScore += finding.Weight
	}
	if len(report.Owners) >= 2 {
		report.RiskScore += 10
		report.Findings = append(report.Findings, Finding{"ownership", "change crosses multiple ownership areas", 10})
	}
	if report.RiskScore > 100 {
		report.RiskScore = 100
	}
	report.Risk = classify(report.RiskScore)
	return report, nil
}

func newServiceImpact(service servicemap.Service, role, contract string) ServiceImpact {
	owners := append([]string(nil), service.Owners...)
	sort.Strings(owners)
	return ServiceImpact{
		Name:        service.Name,
		Role:        role,
		Contract:    contract,
		Criticality: strings.ToLower(strings.TrimSpace(service.Criticality)),
		Owners:      owners,
	}
}

func serviceImpactKey(impact ServiceImpact) string {
	return impact.Contract + "\x00" + impact.Role + "\x00" + impact.Name
}

func serviceRoleRank(role string) int {
	if role == "provider" {
		return 0
	}
	return 1
}

func changedFiles(base, head string) ([]string, error) {
	out, err := exec.Command("git", "diff", "--name-only", base+"..."+head).Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	sort.Strings(lines)
	return lines, nil
}

func classify(score int) string {
	switch {
	case score >= 70:
		return "CRITICAL"
	case score >= 40:
		return "HIGH"
	case score >= 20:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func isOpenAPI(p string) bool {
	b := filepath.Base(p)
	return strings.Contains(p, "openapi") || strings.Contains(p, "swagger") || b == "api.yaml" || b == "api.yml"
}

func isMigration(p string) bool {
	return strings.Contains(p, "migration") || strings.Contains(p, "migrations/") || strings.Contains(p, "db/migrate")
}

func isDeployment(p string) bool {
	return strings.Contains(p, "terraform") || strings.HasSuffix(p, ".tf") || strings.Contains(p, "k8s/") || strings.Contains(p, "kubernetes/") || strings.Contains(p, "helm/") || strings.Contains(p, "dockerfile") || strings.Contains(p, "docker-compose")
}

func isWorkflow(p string) bool {
	return strings.Contains(p, ".github/workflows/") || strings.Contains(p, "gitlab-ci") || strings.Contains(p, "jenkinsfile")
}

func isConfig(p string) bool {
	b := strings.ToLower(filepath.Base(p))
	ext := strings.ToLower(filepath.Ext(b))

	if b == ".impactctl.yml" || b == ".impactctl.yaml" {
		return true
	}
	if b == ".env" || strings.HasPrefix(b, ".env.") || strings.HasSuffix(b, ".env") {
		return true
	}
	if ext == ".toml" || ext == ".ini" {
		return true
	}
	if ext == ".yaml" || ext == ".yml" || ext == ".json" {
		return strings.Contains(b, "config") || strings.Contains(b, "settings") || strings.Contains(b, "application")
	}
	return b == "config" || b == "configuration"
}

func discoverCodeowners(root string) (string, bool) {
	candidates := []string{
		filepath.Join(root, ".github", "CODEOWNERS"),
		filepath.Join(root, "CODEOWNERS"),
		filepath.Join(root, "docs", "CODEOWNERS"),
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, true
		}
	}
	return "", false
}

type codeowners []ownerRule

type ownerRule struct {
	pattern string
	owners  []string
}

func loadCodeowners(path string) (codeowners, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var rules codeowners
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		rules = append(rules, ownerRule{pattern: fields[0], owners: fields[1:]})
	}
	return rules, scanner.Err()
}

func (c codeowners) match(path string) []string {
	var result []string
	for _, rule := range c {
		pattern := strings.TrimPrefix(rule.pattern, "/")
		matched := false
		if strings.HasSuffix(pattern, "/") {
			matched = strings.HasPrefix(path, pattern)
		} else if ok, _ := filepath.Match(pattern, path); ok {
			matched = true
		} else if strings.Contains(pattern, "/") && strings.HasPrefix(path, strings.TrimSuffix(pattern, "*")) {
			matched = true
		} else if !strings.Contains(pattern, "/") {
			matched = filepath.Base(path) == pattern
		}
		if matched {
			result = rule.owners
		}
	}
	return result
}
