package servicemap

import (
	"sort"
	"strings"
)

type DownstreamImpact struct {
	Source  Service
	Service Service
	Path    []string
}

type traversalNode struct {
	name   string
	source string
	path   []string
}

func (c Config) DownstreamFrom(sourceNames []string) []DownstreamImpact {
	services := make(map[string]Service, len(c.Services))
	for _, service := range c.Services {
		services[strings.TrimSpace(service.Name)] = service
	}

	reverse := make(map[string]map[string]struct{}, len(c.Services))
	addEdge := func(upstream, downstream string) {
		upstream = strings.TrimSpace(upstream)
		downstream = strings.TrimSpace(downstream)
		if upstream == "" || downstream == "" || upstream == downstream {
			return
		}
		if reverse[upstream] == nil {
			reverse[upstream] = map[string]struct{}{}
		}
		reverse[upstream][downstream] = struct{}{}
	}

	for _, service := range c.Services {
		dependent := strings.TrimSpace(service.Name)
		for _, dependency := range service.DependsOn {
			addEdge(dependency, dependent)
		}
		provider := dependent
		for _, contract := range service.OpenAPI {
			for _, consumer := range contract.Consumers {
				addEdge(provider, consumer)
			}
		}
	}

	adjacency := make(map[string][]string, len(reverse))
	for upstream, set := range reverse {
		for downstream := range set {
			adjacency[upstream] = append(adjacency[upstream], downstream)
		}
		sort.Strings(adjacency[upstream])
	}

	sourceSet := map[string]struct{}{}
	var sources []string
	for _, source := range sourceNames {
		source = strings.TrimSpace(source)
		if _, ok := services[source]; !ok {
			continue
		}
		if _, exists := sourceSet[source]; exists {
			continue
		}
		sourceSet[source] = struct{}{}
		sources = append(sources, source)
	}
	sort.Strings(sources)

	visited := make(map[string]struct{}, len(services))
	queue := make([]traversalNode, 0, len(services))
	for _, source := range sources {
		visited[source] = struct{}{}
		queue = append(queue, traversalNode{name: source, source: source, path: []string{source}})
	}

	var impacts []DownstreamImpact
	for i := 0; i < len(queue); i++ {
		current := queue[i]
		for _, downstream := range adjacency[current.name] {
			if _, seen := visited[downstream]; seen {
				continue
			}
			service, ok := services[downstream]
			if !ok {
				continue
			}
			visited[downstream] = struct{}{}
			path := append(append([]string(nil), current.path...), downstream)
			impacts = append(impacts, DownstreamImpact{
				Source:  services[current.source],
				Service: service,
				Path:    path,
			})
			queue = append(queue, traversalNode{name: downstream, source: current.source, path: path})
		}
	}

	sort.Slice(impacts, func(i, j int) bool {
		if len(impacts[i].Path) != len(impacts[j].Path) {
			return len(impacts[i].Path) < len(impacts[j].Path)
		}
		left := strings.Join(impacts[i].Path, "\x00")
		right := strings.Join(impacts[j].Path, "\x00")
		return left < right
	})
	return impacts
}
