package ecosystems

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLookupNormalizesDirectDependenciesWithProvenance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/hosts/GitHub/repositories/acme%2Fdemo/manifests" &&
			r.URL.Path != "/api/v1/hosts/GitHub/repositories/acme/demo/manifests" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"filepath":"go.mod",
			"repository_link":"https://github.com/acme/demo/blob/main/go.mod",
			"dependencies":[
				{"package_name":"example/direct","ecosystem":"go","direct":true},
				{"package_name":"example/transitive","ecosystem":"go","direct":false}
			]
		}]`))
	}))
	defer server.Close()

	fixed := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	client := NewClient()
	client.BaseURL = server.URL
	client.Now = func() time.Time { return fixed }

	got, err := client.Lookup(context.Background(), "acme/demo")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Facts) != 1 {
		t.Fatalf("facts = %d, want 1 direct fact", len(got.Facts))
	}
	fact := got.Facts[0]
	if fact.Kind != "declared_dependency" || fact.Object != "go:example/direct" || !fact.Direct {
		t.Fatalf("unexpected fact: %#v", fact)
	}
	if fact.Evidence.Manifest != "go.mod" || fact.Evidence.SourceURL == "" {
		t.Fatalf("missing provenance: %#v", fact.Evidence)
	}
	if !got.RetrievedAt.Equal(fixed) {
		t.Fatalf("retrieved_at = %s", got.RetrievedAt)
	}
}

func TestLookupFailsClosedOnAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewClient()
	client.BaseURL = server.URL
	if _, err := client.Lookup(context.Background(), "acme/demo"); err == nil {
		t.Fatal("expected API error")
	}
}
