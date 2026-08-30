package servicemap

import (
	"reflect"
	"strings"
	"testing"
)

func TestOpenAPIImpactsForPath(t *testing.T) {
	cfg, err := Parse([]byte(`version: 1
services:
  - name: orders
    paths: [services/orders/**]
    criticality: high
    owners: ['@orders-team']
    openapi:
      - path: api/orders/openapi.yaml
        consumers: [checkout, mobile]
  - name: checkout
    paths: [services/checkout/**]
    owners: ['@checkout-team']
  - name: mobile
    paths: [apps/mobile/**]
    owners: ['@mobile-team']
`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	impacts := cfg.OpenAPIImpactsForPath("api/orders/openapi.yaml")
	if len(impacts) != 1 {
		t.Fatalf("OpenAPIImpactsForPath() len = %d, want 1: %#v", len(impacts), impacts)
	}
	if impacts[0].Provider.Name != "orders" {
		t.Fatalf("provider = %q, want orders", impacts[0].Provider.Name)
	}
	var consumers []string
	for _, service := range impacts[0].Consumers {
		consumers = append(consumers, service.Name)
	}
	if want := []string{"checkout", "mobile"}; !reflect.DeepEqual(consumers, want) {
		t.Fatalf("consumers = %v, want %v", consumers, want)
	}
}

func TestOpenAPIImpactsRequireExplicitMapping(t *testing.T) {
	cfg, err := Parse([]byte(`version: 1
services:
  - name: orders
    paths: [services/orders/**]
  - name: checkout
    paths: [services/checkout/**]
`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := cfg.OpenAPIImpactsForPath("api/orders/openapi.yaml"); len(got) != 0 {
		t.Fatalf("OpenAPIImpactsForPath() = %#v, want no inferred relationship", got)
	}
}

func TestOpenAPIConfigRejectsUnknownConsumer(t *testing.T) {
	_, err := Parse([]byte(`version: 1
services:
  - name: orders
    paths: [services/orders/**]
    openapi:
      - path: api/orders/openapi.yaml
        consumers: [missing]
`))
	if err == nil || !strings.Contains(err.Error(), "unknown consumer") {
		t.Fatalf("Parse() error = %v, want unknown consumer error", err)
	}
}

func TestOpenAPIConfigRejectsSelfConsumer(t *testing.T) {
	_, err := Parse([]byte(`version: 1
services:
  - name: orders
    paths: [services/orders/**]
    openapi:
      - path: api/orders/openapi.yaml
        consumers: [orders]
`))
	if err == nil || !strings.Contains(err.Error(), "cannot consume itself") {
		t.Fatalf("Parse() error = %v, want self-consumer error", err)
	}
}
