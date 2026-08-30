package servicemap

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseValidConfig(t *testing.T) {
	cfg, err := Parse([]byte(`version: 1
services:
  - name: checkout
    paths:
      - services/checkout/**
      - api/checkout/*
    criticality: high
    owners:
      - "@checkout-team"
  - name: inventory
    paths:
      - services/inventory
    criticality: medium
    owners:
      - "@inventory-team"
`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if cfg.Version != 1 || len(cfg.Services) != 2 {
		t.Fatalf("unexpected config: %#v", cfg)
	}
	if got := cfg.Services[0].Criticality; got != "high" {
		t.Fatalf("criticality = %q, want high", got)
	}
}

func TestParseRejectsUnknownField(t *testing.T) {
	_, err := Parse([]byte(`version: 1
services:
  - name: checkout
    paths: [services/checkout/**]
    mystery: true
`))
	if err == nil || !strings.Contains(err.Error(), "field mystery not found") {
		t.Fatalf("Parse() error = %v, want unknown-field error", err)
	}
}

func TestValidateRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "version",
			yaml: "version: 2\nservices:\n  - name: api\n    paths: [services/api/**]\n",
			want: "unsupported schema version",
		},
		{
			name: "duplicate service",
			yaml: "version: 1\nservices:\n  - name: api\n    paths: [services/api/**]\n  - name: api\n    paths: [services/other/**]\n",
			want: "duplicate service name",
		},
		{
			name: "missing paths",
			yaml: "version: 1\nservices:\n  - name: api\n",
			want: "must declare at least one path",
		},
		{
			name: "repository escape",
			yaml: "version: 1\nservices:\n  - name: api\n    paths: [../api]\n",
			want: "must not escape the repository",
		},
		{
			name: "criticality",
			yaml: "version: 1\nservices:\n  - name: api\n    paths: [services/api/**]\n    criticality: urgent\n",
			want: "invalid criticality",
		},
		{
			name: "duplicate owner",
			yaml: "version: 1\nservices:\n  - name: api\n    paths: [services/api/**]\n    owners: ['@team', '@team']\n",
			want: "duplicate owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.yaml))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Parse() error = %v, want substring %q", err, tt.want)
			}
		})
	}
}

func TestServicesForPath(t *testing.T) {
	cfg := Config{
		Version: 1,
		Services: []Service{
			{Name: "payments", Paths: []string{"services/payments/**", "shared/*.go"}},
			{Name: "checkout", Paths: []string{"services/checkout"}},
			{Name: "shared", Paths: []string{"shared/**"}},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	matches := cfg.ServicesForPath("shared/client.go")
	var names []string
	for _, service := range matches {
		names = append(names, service.Name)
	}
	want := []string{"payments", "shared"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("ServicesForPath() = %v, want %v", names, want)
	}

	matches = cfg.ServicesForPath("services/checkout/handler.go")
	if len(matches) != 1 || matches[0].Name != "checkout" {
		t.Fatalf("checkout match = %#v", matches)
	}
}

func TestLoadOptional(t *testing.T) {
	root := t.TempDir()
	if _, exists, err := Load(root); err != nil || exists {
		t.Fatalf("Load(absent) exists=%v err=%v", exists, err)
	}

	data := `version: 1
services:
  - name: api
    paths:
      - services/api/**
`
	if err := os.WriteFile(filepath.Join(root, Filename), []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, exists, err := Load(root)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !exists || len(cfg.Services) != 1 || cfg.Services[0].Name != "api" {
		t.Fatalf("Load() = %#v, exists=%v", cfg, exists)
	}
}

func TestParseRejectsMultipleDocuments(t *testing.T) {
	_, err := Parse([]byte("version: 1\nservices:\n  - name: api\n    paths: [services/api/**]\n---\nversion: 1\nservices:\n  - name: other\n    paths: [services/other/**]\n"))
	if err == nil || !strings.Contains(err.Error(), "multiple YAML documents") {
		t.Fatalf("Parse() error = %v, want multiple document error", err)
	}
}
