package impact

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	AsyncAPIAdditive = "ADDITIVE"
	AsyncAPIBreaking = "BREAKING"
	AsyncAPIReview   = "REVIEW"
)

type AsyncAPIImpact struct {
	Path   string
	Kind   string
	Name   string
	Change string
	Detail string
}

type asyncAPIDocument struct {
	AsyncAPI   string                 `yaml:"asyncapi"`
	Channels   map[string]any         `yaml:"channels"`
	Components asyncAPIComponents     `yaml:"components"`
}

type asyncAPIComponents struct {
	Messages map[string]any `yaml:"messages"`
	Schemas  map[string]any `yaml:"schemas"`
}

func analyzeAsyncAPI(base, head, path string) ([]AsyncAPIImpact, bool, error) {
	if !isAsyncAPIPath(path) {
		return nil, false, nil
	}

	baseData, baseExists, err := gitFileAtRef(base, path)
	if err != nil {
		return nil, false, err
	}
	headData, headExists, err := gitFileAtRef(head, path)
	if err != nil {
		return nil, false, err
	}

	var before, after asyncAPIDocument
	beforeAsync, afterAsync := false, false
	if baseExists {
		if err := yaml.Unmarshal(baseData, &before); err != nil {
			return nil, false, fmt.Errorf("parse AsyncAPI base %s: %w", path, err)
		}
		beforeAsync = strings.TrimSpace(before.AsyncAPI) != ""
	}
	if headExists {
		if err := yaml.Unmarshal(headData, &after); err != nil {
			return nil, false, fmt.Errorf("parse AsyncAPI head %s: %w", path, err)
		}
		afterAsync = strings.TrimSpace(after.AsyncAPI) != ""
	}

	if !beforeAsync && !afterAsync {
		return nil, false, nil
	}

	var impacts []AsyncAPIImpact
	impacts = append(impacts, diffAsyncAPIEntities(path, "channel", before.Channels, after.Channels)...)
	impacts = append(impacts, diffAsyncAPIEntities(path, "message", before.Components.Messages, after.Components.Messages)...)
	impacts = append(impacts, diffAsyncAPIEntities(path, "schema", before.Components.Schemas, after.Components.Schemas)...)

	if len(impacts) == 0 && !reflect.DeepEqual(before, after) {
		impacts = append(impacts, AsyncAPIImpact{
			Path:   path,
			Kind:   "contract",
			Name:   path,
			Change: AsyncAPIReview,
			Detail: "AsyncAPI document changed outside safely classified channels/messages/schemas",
		})
	}

	sort.Slice(impacts, func(i, j int) bool {
		if asyncAPIChangeRank(impacts[i].Change) != asyncAPIChangeRank(impacts[j].Change) {
			return asyncAPIChangeRank(impacts[i].Change) < asyncAPIChangeRank(impacts[j].Change)
		}
		if impacts[i].Kind != impacts[j].Kind {
			return impacts[i].Kind < impacts[j].Kind
		}
		return impacts[i].Name < impacts[j].Name
	})
	return impacts, true, nil
}

func diffAsyncAPIEntities(path, kind string, before, after map[string]any) []AsyncAPIImpact {
	if before == nil {
		before = map[string]any{}
	}
	if after == nil {
		after = map[string]any{}
	}

	var impacts []AsyncAPIImpact
	for name, beforeValue := range before {
		afterValue, exists := after[name]
		if !exists {
			impacts = append(impacts, AsyncAPIImpact{
				Path: path, Kind: kind, Name: name, Change: AsyncAPIBreaking,
				Detail: fmt.Sprintf("%s %q was removed", kind, name),
			})
			continue
		}
		if !reflect.DeepEqual(beforeValue, afterValue) {
			impacts = append(impacts, AsyncAPIImpact{
				Path: path, Kind: kind, Name: name, Change: AsyncAPIReview,
				Detail: fmt.Sprintf("%s %q changed and needs semantic review", kind, name),
			})
		}
	}
	for name := range after {
		if _, exists := before[name]; !exists {
			impacts = append(impacts, AsyncAPIImpact{
				Path: path, Kind: kind, Name: name, Change: AsyncAPIAdditive,
				Detail: fmt.Sprintf("%s %q was added", kind, name),
			})
		}
	}
	return impacts
}

func asyncAPIFinding(path string, impacts []AsyncAPIImpact) Finding {
	change := AsyncAPIAdditive
	for _, impact := range impacts {
		if impact.Change == AsyncAPIBreaking {
			change = AsyncAPIBreaking
			break
		}
		if impact.Change == AsyncAPIReview {
			change = AsyncAPIReview
		}
	}

	switch change {
	case AsyncAPIBreaking:
		return Finding{"event-breaking", path + " removes or renames AsyncAPI contract entities", 35}
	case AsyncAPIReview:
		return Finding{"event-review", path + " changes AsyncAPI semantics that require review", 15}
	default:
		return Finding{"event-additive", path + " adds AsyncAPI contract entities", 5}
	}
}

func asyncAPIChangeRank(change string) int {
	switch change {
	case AsyncAPIBreaking:
		return 0
	case AsyncAPIReview:
		return 1
	default:
		return 2
	}
}

func isAsyncAPIPath(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	ext := strings.ToLower(filepath.Ext(base))
	if ext != ".yaml" && ext != ".yml" && ext != ".json" {
		return false
	}
	return strings.Contains(base, "asyncapi")
}

func gitFileAtRef(ref, path string) ([]byte, bool, error) {
	cmd := exec.Command("git", "show", ref+":"+path)
	out, err := cmd.Output()
	if err == nil {
		return out, true, nil
	}
	if _, ok := err.(*exec.ExitError); ok {
		return nil, false, nil
	}
	return nil, false, fmt.Errorf("git show %s:%s failed: %w", ref, path, err)
}
