package sapimpact

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const SchemaVersion = 1

type Manifest struct {
	Version int    `yaml:"version" json:"version"`
	Change  Change `yaml:"change" json:"change"`
	Nodes   []Node `yaml:"nodes" json:"nodes"`
}

type Change struct {
	ID          string   `yaml:"id" json:"id"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Changed     []string `yaml:"changed" json:"changed"`
}

type Node struct {
	Name        string   `yaml:"name" json:"name"`
	Kind        string   `yaml:"kind" json:"kind"`
	Criticality string   `yaml:"criticality,omitempty" json:"criticality,omitempty"`
	Owners      []string `yaml:"owners,omitempty" json:"owners,omitempty"`
	DependsOn   []string `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
}

type Impact struct {
	Name        string   `json:"name"`
	Kind        string   `json:"kind"`
	Criticality string   `json:"criticality,omitempty"`
	Owners      []string `json:"owners,omitempty"`
	Path        []string `json:"path"`
}

type Report struct {
	ChangeID            string   `json:"change_id"`
	Description         string   `json:"description,omitempty"`
	Risk                string   `json:"risk"`
	RiskScore           int      `json:"risk_score"`
	Changed             []Node   `json:"changed"`
	Downstream          []Impact `json:"downstream"`
	AffectedProcesses   []string `json:"affected_processes,omitempty"`
	SuggestedReviewers  []string `json:"suggested_reviewers,omitempty"`
}

type traversal struct {
	name string
	path []string
}

var allowedKinds = map[string]struct{}{
	"sap-object":       {},
	"sap-system":       {},
	"integration":      {},
	"api":              {},
	"event":            {},
	"application":      {},
	"data":             {},
	"job":              {},
	"business-process": {},
}

func Load(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read SAP impact manifest: %w", err)
	}
	return Parse(data)
}

func Parse(data []byte) (Manifest, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse SAP impact manifest: %w", err)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return Manifest{}, fmt.Errorf("parse SAP impact manifest: %w", err)
		}
		return Manifest{}, fmt.Errorf("multiple YAML documents are not supported")
	}

	if err := manifest.Validate(); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func (m Manifest) Validate() error {
	if m.Version != SchemaVersion {
		return fmt.Errorf("unsupported SAP impact schema version %d; expected %d", m.Version, SchemaVersion)
	}
	if strings.TrimSpace(m.Change.ID) == "" {
		return fmt.Errorf("change.id must not be empty")
	}
	if len(m.Change.Changed) == 0 {
		return fmt.Errorf("change.changed must contain at least one node")
	}
	if len(m.Nodes) == 0 {
		return fmt.Errorf("nodes must contain at least one node")
	}

	nodes := make(map[string]Node, len(m.Nodes))
	for i, node := range m.Nodes {
		name := strings.TrimSpace(node.Name)
		if name == "" {
			return fmt.Errorf("nodes[%d].name must not be empty", i)
		}
		if _, exists := nodes[name]; exists {
			return fmt.Errorf("duplicate node name %q", name)
		}

		kind := strings.ToLower(strings.TrimSpace(node.Kind))
		if _, ok := allowedKinds[kind]; !ok {
			return fmt.Errorf("node %q has unsupported kind %q", name, node.Kind)
		}

		criticality := strings.ToLower(strings.TrimSpace(node.Criticality))
		switch criticality {
		case "", "low", "medium", "high", "critical":
		default:
			return fmt.Errorf("node %q has invalid criticality %q", name, node.Criticality)
		}

		if err := validateUniqueNonEmpty(node.Owners, fmt.Sprintf("node %q owners", name)); err != nil {
			return err
		}
		if err := validateUniqueNonEmpty(node.DependsOn, fmt.Sprintf("node %q dependencies", name)); err != nil {
			return err
		}
		nodes[name] = node
	}

	seenChanged := map[string]struct{}{}
	for _, changed := range m.Change.Changed {
		changed = strings.TrimSpace(changed)
		if changed == "" {
			return fmt.Errorf("change.changed contains an empty node")
		}
		if _, ok := nodes[changed]; !ok {
			return fmt.Errorf("change.changed references unknown node %q", changed)
		}
		if _, exists := seenChanged[changed]; exists {
			return fmt.Errorf("change.changed contains duplicate node %q", changed)
		}
		seenChanged[changed] = struct{}{}
	}

	for _, node := range m.Nodes {
		name := strings.TrimSpace(node.Name)
		for _, dependency := range node.DependsOn {
			dependency = strings.TrimSpace(dependency)
			if dependency == name {
				return fmt.Errorf("node %q cannot depend on itself", name)
			}
			if _, ok := nodes[dependency]; !ok {
				return fmt.Errorf("node %q references unknown dependency %q", name, dependency)
			}
		}
	}
	return nil
}

func Analyze(m Manifest) Report {
	nodes := make(map[string]Node, len(m.Nodes))
	for _, node := range m.Nodes {
		node.Name = strings.TrimSpace(node.Name)
		node.Kind = strings.ToLower(strings.TrimSpace(node.Kind))
		node.Criticality = strings.ToLower(strings.TrimSpace(node.Criticality))
		nodes[node.Name] = node
	}

	reverse := make(map[string][]string)
	for _, node := range nodes {
		for _, dependency := range node.DependsOn {
			dependency = strings.TrimSpace(dependency)
			reverse[dependency] = append(reverse[dependency], node.Name)
		}
	}
	for source := range reverse {
		sort.Strings(reverse[source])
	}

	changedNames := append([]string(nil), m.Change.Changed...)
	for i := range changedNames {
		changedNames[i] = strings.TrimSpace(changedNames[i])
	}
	sort.Strings(changedNames)

	changed := make([]Node, 0, len(changedNames))
	visited := make(map[string]struct{}, len(nodes))
	queue := make([]traversal, 0, len(nodes))
	for _, name := range changedNames {
		node := nodes[name]
		changed = append(changed, node)
		visited[name] = struct{}{}
		queue = append(queue, traversal{name: name, path: []string{name}})
	}

	var downstream []Impact
	for i := 0; i < len(queue); i++ {
		current := queue[i]
		for _, next := range reverse[current.name] {
			if _, seen := visited[next]; seen {
				continue
			}
			node := nodes[next]
			visited[next] = struct{}{}
			path := append(append([]string(nil), current.path...), next)
			downstream = append(downstream, Impact{
				Name:        node.Name,
				Kind:        node.Kind,
				Criticality: node.Criticality,
				Owners:      sortedCopy(node.Owners),
				Path:        path,
			})
			queue = append(queue, traversal{name: next, path: path})
		}
	}

	sort.Slice(downstream, func(i, j int) bool {
		if len(downstream[i].Path) != len(downstream[j].Path) {
			return len(downstream[i].Path) < len(downstream[j].Path)
		}
		return strings.Join(downstream[i].Path, "\x00") < strings.Join(downstream[j].Path, "\x00")
	})

	owners := map[string]struct{}{}
	processes := map[string]struct{}{}
	maxCriticality := ""
	kinds := map[string]struct{}{}
	for _, node := range changed {
		collectOwners(owners, node.Owners)
		maxCriticality = higherCriticality(maxCriticality, node.Criticality)
		kinds[node.Kind] = struct{}{}
	}
	for _, impact := range downstream {
		collectOwners(owners, impact.Owners)
		maxCriticality = higherCriticality(maxCriticality, impact.Criticality)
		kinds[impact.Kind] = struct{}{}
		if impact.Kind == "business-process" {
			processes[impact.Name] = struct{}{}
		}
	}

	score := 10
	score += min(len(downstream)*8, 32)
	score += len(processes) * 15
	score += criticalityWeight(maxCriticality)
	if len(kinds) >= 4 {
		score += 5
	}
	if score > 100 {
		score = 100
	}

	return Report{
		ChangeID:           strings.TrimSpace(m.Change.ID),
		Description:        strings.TrimSpace(m.Change.Description),
		Risk:               riskLabel(score),
		RiskScore:          score,
		Changed:            changed,
		Downstream:         downstream,
		AffectedProcesses:  sortedKeys(processes),
		SuggestedReviewers: sortedKeys(owners),
	}
}

func validateUniqueNonEmpty(values []string, field string) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("%s contains an empty value", field)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("%s contains duplicate value %q", field, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func collectOwners(target map[string]struct{}, owners []string) {
	for _, owner := range owners {
		owner = strings.TrimSpace(owner)
		if owner != "" {
			target[owner] = struct{}{}
		}
	}
}

func sortedCopy(values []string) []string {
	result := append([]string(nil), values...)
	for i := range result {
		result[i] = strings.TrimSpace(result[i])
	}
	sort.Strings(result)
	return result
}

func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func higherCriticality(a, b string) string {
	if criticalityWeight(b) > criticalityWeight(a) {
		return strings.ToLower(strings.TrimSpace(b))
	}
	return strings.ToLower(strings.TrimSpace(a))
}

func criticalityWeight(value string) int {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "medium":
		return 10
	case "high":
		return 20
	case "critical":
		return 30
	default:
		return 0
	}
}

func riskLabel(score int) string {
	switch {
	case score >= 75:
		return "CRITICAL"
	case score >= 50:
		return "HIGH"
	case score >= 25:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
