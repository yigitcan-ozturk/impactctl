package servicemap

import (
	"bytes"
	"fmt"
	"io"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	Filename      = ".impactctl.yml"
	SchemaVersion = 1
)

type Config struct {
	Version  int       `yaml:"version"`
	Services []Service `yaml:"services"`
}

type Service struct {
	Name        string            `yaml:"name"`
	Paths       []string          `yaml:"paths"`
	Criticality string            `yaml:"criticality,omitempty"`
	Owners      []string          `yaml:"owners,omitempty"`
	DependsOn   []string          `yaml:"depends_on,omitempty"`
	OpenAPI     []OpenAPIContract `yaml:"openapi,omitempty"`
}

func Load(root string) (Config, bool, error) {
	configPath := filepath.Join(root, Filename)
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, false, nil
		}
		return Config{}, false, fmt.Errorf("read %s: %w", Filename, err)
	}

	cfg, err := Parse(data)
	if err != nil {
		return Config{}, true, fmt.Errorf("parse %s: %w", Filename, err)
	}
	return cfg, true, nil
}

func Parse(data []byte) (Config, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return Config{}, err
		}
		return Config{}, fmt.Errorf("multiple YAML documents are not supported")
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.Version != SchemaVersion {
		return fmt.Errorf("unsupported schema version %d; expected %d", c.Version, SchemaVersion)
	}
	if len(c.Services) == 0 {
		return fmt.Errorf("services must contain at least one service")
	}

	names := make(map[string]struct{}, len(c.Services))
	for i, service := range c.Services {
		name := strings.TrimSpace(service.Name)
		if name == "" {
			return fmt.Errorf("services[%d].name must not be empty", i)
		}
		if _, exists := names[name]; exists {
			return fmt.Errorf("duplicate service name %q", name)
		}
		names[name] = struct{}{}
	}

	for _, service := range c.Services {
		name := strings.TrimSpace(service.Name)
		if len(service.Paths) == 0 {
			return fmt.Errorf("service %q must declare at least one path", name)
		}
		seenPaths := map[string]struct{}{}
		for _, pattern := range service.Paths {
			pattern = strings.TrimSpace(pattern)
			if err := validatePattern(pattern); err != nil {
				return fmt.Errorf("service %q path %q: %w", name, pattern, err)
			}
			if _, exists := seenPaths[pattern]; exists {
				return fmt.Errorf("service %q contains duplicate path %q", name, pattern)
			}
			seenPaths[pattern] = struct{}{}
		}

		criticality := strings.ToLower(strings.TrimSpace(service.Criticality))
		switch criticality {
		case "", "low", "medium", "high", "critical":
		default:
			return fmt.Errorf("service %q has invalid criticality %q", name, service.Criticality)
		}

		seenOwners := map[string]struct{}{}
		for _, owner := range service.Owners {
			owner = strings.TrimSpace(owner)
			if owner == "" {
				return fmt.Errorf("service %q contains an empty owner", name)
			}
			if _, exists := seenOwners[owner]; exists {
				return fmt.Errorf("service %q contains duplicate owner %q", name, owner)
			}
			seenOwners[owner] = struct{}{}
		}

		seenDependencies := map[string]struct{}{}
		for _, dependency := range service.DependsOn {
			dependency = strings.TrimSpace(dependency)
			if dependency == "" {
				return fmt.Errorf("service %q contains an empty dependency", name)
			}
			if dependency == name {
				return fmt.Errorf("service %q cannot depend on itself", name)
			}
			if _, exists := names[dependency]; !exists {
				return fmt.Errorf("service %q references unknown dependency %q", name, dependency)
			}
			if _, exists := seenDependencies[dependency]; exists {
				return fmt.Errorf("service %q contains duplicate dependency %q", name, dependency)
			}
			seenDependencies[dependency] = struct{}{}
		}

		seenContracts := map[string]struct{}{}
		for _, contract := range service.OpenAPI {
			contractPath := strings.TrimSpace(contract.Path)
			if err := validatePattern(contractPath); err != nil {
				return fmt.Errorf("service %q OpenAPI path %q: %w", name, contractPath, err)
			}
			if _, exists := seenContracts[contractPath]; exists {
				return fmt.Errorf("service %q contains duplicate OpenAPI path %q", name, contractPath)
			}
			seenContracts[contractPath] = struct{}{}

			seenConsumers := map[string]struct{}{}
			for _, consumer := range contract.Consumers {
				consumer = strings.TrimSpace(consumer)
				if consumer == "" {
					return fmt.Errorf("service %q OpenAPI path %q contains an empty consumer", name, contractPath)
				}
				if consumer == name {
					return fmt.Errorf("service %q OpenAPI path %q cannot consume itself", name, contractPath)
				}
				if _, exists := names[consumer]; !exists {
					return fmt.Errorf("service %q OpenAPI path %q references unknown consumer %q", name, contractPath, consumer)
				}
				if _, exists := seenConsumers[consumer]; exists {
					return fmt.Errorf("service %q OpenAPI path %q contains duplicate consumer %q", name, contractPath, consumer)
				}
				seenConsumers[consumer] = struct{}{}
			}
		}
	}
	return nil
}

func (c Config) ServicesForPath(repoPath string) []Service {
	candidate := normalizeRepoPath(repoPath)
	var matches []Service
	for _, service := range c.Services {
		for _, pattern := range service.Paths {
			if matchPattern(pattern, candidate) {
				matches = append(matches, service)
				break
			}
		}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})
	return matches
}

func validatePattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("must not be empty")
	}
	if strings.Contains(pattern, "\\") {
		return fmt.Errorf("must use forward slashes")
	}
	if strings.HasPrefix(pattern, "/") || filepath.IsAbs(pattern) {
		return fmt.Errorf("must be repository-relative")
	}

	cleaned := pathpkg.Clean(strings.TrimPrefix(pattern, "./"))
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return fmt.Errorf("must not escape the repository")
	}

	if strings.ContainsAny(pattern, "*?[") {
		probe := pattern
		if strings.HasSuffix(probe, "/**") {
			probe = strings.TrimSuffix(probe, "/**") + "/*"
		}
		if _, err := pathpkg.Match(probe, "probe"); err != nil {
			return fmt.Errorf("invalid glob: %w", err)
		}
	}
	return nil
}

func matchPattern(pattern, candidate string) bool {
	pattern = normalizeRepoPath(pattern)
	candidate = normalizeRepoPath(candidate)

	if strings.HasSuffix(pattern, "/**") {
		root := strings.TrimSuffix(pattern, "/**")
		return candidate == root || strings.HasPrefix(candidate, root+"/")
	}

	if strings.ContainsAny(pattern, "*?[") {
		matched, err := pathpkg.Match(pattern, candidate)
		return err == nil && matched
	}

	return candidate == pattern || strings.HasPrefix(candidate, strings.TrimSuffix(pattern, "/")+"/")
}

func normalizeRepoPath(value string) string {
	value = strings.TrimSpace(filepath.ToSlash(value))
	value = strings.TrimPrefix(value, "./")
	return pathpkg.Clean(value)
}
