package ecosystems

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://repos.ecosyste.ms"

type Evidence struct {
	Source      string     `json:"source"`
	SourceType  string     `json:"source_type"`
	Repository  string     `json:"repository"`
	RetrievedAt time.Time  `json:"retrieved_at"`
	Facts       []Fact     `json:"facts"`
}

type Fact struct {
	Kind     string       `json:"kind"`
	Subject  string       `json:"subject"`
	Object   string       `json:"object"`
	Direct   bool         `json:"direct"`
	Kind     string       `json:"dependency_kind,omitempty"`
	Optional bool         `json:"optional"`
	Evidence FactEvidence `json:"evidence"`
}

type FactEvidence struct {
	Manifest  string `json:"manifest"`
	SourceURL string `json:"source_url"`
}

type manifest struct {
	Filepath       string       `json:"filepath"`
	RepositoryLink string       `json:"repository_link"`
	Dependencies   []dependency `json:"dependencies"`
}

type dependency struct {
	PackageName string `json:"package_name"`
	Ecosystem   string `json:"ecosystem"`
	Direct      bool   `json:"direct"`
	Kind        string `json:"kind"`
	Optional    bool   `json:"optional"`
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Now        func() time.Time
}

func NewClient() *Client {
	return &Client{
		BaseURL: defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		Now: time.Now,
	}
}

func (c *Client) Lookup(ctx context.Context, repository string) (Evidence, error) {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return Evidence{}, fmt.Errorf("ecosystems: repository must be owner/name")
	}

	base := strings.TrimRight(c.BaseURL, "/")
	endpoint := fmt.Sprintf("%s/api/v1/hosts/GitHub/repositories/%s/manifests", base, url.PathEscape(repository))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Evidence{}, fmt.Errorf("ecosystems: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "impactctl/experimental-ecosystems-adapter")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return Evidence{}, fmt.Errorf("ecosystems: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Evidence{}, fmt.Errorf("ecosystems: API returned %s", resp.Status)
	}

	var manifests []manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifests); err != nil {
		return Evidence{}, fmt.Errorf("ecosystems: decode response: %w", err)
	}

	out := Evidence{
		Source: "ecosyste.ms",
		SourceType: "external_repository_manifest",
		Repository: repository,
		RetrievedAt: c.Now().UTC(),
		Facts: []Fact{},
	}
	for _, m := range manifests {
		for _, d := range m.Dependencies {
			if !d.Direct || strings.TrimSpace(d.PackageName) == "" {
				continue
			}
			object := d.PackageName
			if d.Ecosystem != "" {
				object = d.Ecosystem + ":" + d.PackageName
			}
			out.Facts = append(out.Facts, Fact{
				Kind: "declared_dependency",
				Subject: repository,
				Object: object,
				Direct: true,
				Kind: d.Kind,
				Optional: d.Optional,
				Evidence: FactEvidence{Manifest: m.Filepath, SourceURL: m.RepositoryLink},
			})
		}
	}
	return out, nil
}
