package servicemap

import (
	"reflect"
	"strings"
	"testing"
)

func TestDownstreamFromLinearGraph(t *testing.T) {
	cfg := mustParseServiceMap(t, `version: 1
services:
  - name: api
    paths: [services/api/**]
  - name: checkout
    paths: [services/checkout/**]
    depends_on: [api]
  - name: billing
    paths: [services/billing/**]
    depends_on: [checkout]
`)

	got := downstreamPaths(cfg.DownstreamFrom([]string{"api"}))
	want := []string{"api>checkout", "api>checkout>billing"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DownstreamFrom() = %v, want %v", got, want)
	}
}

func TestDownstreamFromBranchingGraphIsDeterministic(t *testing.T) {
	cfg := mustParseServiceMap(t, `version: 1
services:
  - name: api
    paths: [services/api/**]
  - name: checkout
    paths: [services/checkout/**]
    depends_on: [api]
  - name: mobile
    paths: [apps/mobile/**]
    depends_on: [api]
  - name: notifications
    paths: [services/notifications/**]
    depends_on: [mobile, checkout]
`)

	got := downstreamPaths(cfg.DownstreamFrom([]string{"api"}))
	want := []string{"api>checkout", "api>mobile", "api>checkout>notifications"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DownstreamFrom() = %v, want %v", got, want)
	}
}

func TestDownstreamFromCycleTerminatesAndDoesNotReturnSource(t *testing.T) {
	cfg := mustParseServiceMap(t, `version: 1
services:
  - name: a
    paths: [a/**]
    depends_on: [c]
  - name: b
    paths: [b/**]
    depends_on: [a]
  - name: c
    paths: [c/**]
    depends_on: [b]
`)

	got := downstreamPaths(cfg.DownstreamFrom([]string{"a"}))
	want := []string{"a>b", "a>b>c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DownstreamFrom() = %v, want %v", got, want)
	}
}

func TestDownstreamFromIncludesExplicitOpenAPIConsumers(t *testing.T) {
	cfg := mustParseServiceMap(t, `version: 1
services:
  - name: orders
    paths: [services/orders/**]
    openapi:
      - path: api/orders/openapi.yaml
        consumers: [checkout]
  - name: checkout
    paths: [services/checkout/**]
  - name: billing
    paths: [services/billing/**]
    depends_on: [checkout]
`)

	got := downstreamPaths(cfg.DownstreamFrom([]string{"orders"}))
	want := []string{"orders>checkout", "orders>checkout>billing"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DownstreamFrom() = %v, want %v", got, want)
	}
}

func TestDependencyConfigRejectsInvalidEdges(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
		want string
	}{
		{
			name: "unknown",
			yaml: "version: 1\nservices:\n  - name: api\n    paths: [api/**]\n    depends_on: [missing]\n",
			want: "unknown dependency",
		},
		{
			name: "self",
			yaml: "version: 1\nservices:\n  - name: api\n    paths: [api/**]\n    depends_on: [api]\n",
			want: "cannot depend on itself",
		},
		{
			name: "duplicate",
			yaml: "version: 1\nservices:\n  - name: base\n    paths: [base/**]\n  - name: api\n    paths: [api/**]\n    depends_on: [base, base]\n",
			want: "duplicate dependency",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(tc.yaml))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Parse() error = %v, want %q", err, tc.want)
			}
		})
	}
}

func mustParseServiceMap(t *testing.T, data string) Config {
	t.Helper()
	cfg, err := Parse([]byte(data))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	return cfg
}

func downstreamPaths(impacts []DownstreamImpact) []string {
	var paths []string
	for _, impact := range impacts {
		paths = append(paths, strings.Join(impact.Path, ">"))
	}
	return paths
}
