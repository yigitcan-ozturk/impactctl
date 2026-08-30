package servicemap

import "sort"

type OpenAPIContract struct {
	Path      string   `yaml:"path"`
	Consumers []string `yaml:"consumers,omitempty"`
}

type OpenAPIImpact struct {
	Contract  string
	Provider  Service
	Consumers []Service
}

func (c Config) OpenAPIImpactsForPath(repoPath string) []OpenAPIImpact {
	candidate := normalizeRepoPath(repoPath)
	servicesByName := make(map[string]Service, len(c.Services))
	for _, service := range c.Services {
		servicesByName[service.Name] = service
	}

	var impacts []OpenAPIImpact
	for _, provider := range c.Services {
		for _, contract := range provider.OpenAPI {
			if !matchPattern(contract.Path, candidate) {
				continue
			}

			impact := OpenAPIImpact{
				Contract: contract.Path,
				Provider: provider,
			}
			for _, name := range contract.Consumers {
				if consumer, ok := servicesByName[name]; ok {
					impact.Consumers = append(impact.Consumers, consumer)
				}
			}
			sort.Slice(impact.Consumers, func(i, j int) bool {
				return impact.Consumers[i].Name < impact.Consumers[j].Name
			})
			impacts = append(impacts, impact)
		}
	}

	sort.Slice(impacts, func(i, j int) bool {
		if impacts[i].Provider.Name == impacts[j].Provider.Name {
			return impacts[i].Contract < impacts[j].Contract
		}
		return impacts[i].Provider.Name < impacts[j].Provider.Name
	})
	return impacts
}
