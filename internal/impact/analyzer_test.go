package impact

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		score int
		want  string
	}{
		{0, "LOW"}, {19, "LOW"}, {20, "MEDIUM"}, {40, "HIGH"}, {70, "CRITICAL"},
	}
	for _, tc := range cases {
		if got := classify(tc.score); got != tc.want {
			t.Fatalf("classify(%d)=%s want %s", tc.score, got, tc.want)
		}
	}
}

func TestDetectors(t *testing.T) {
	if !isOpenAPI("api/openapi.yaml") {
		t.Fatal("expected OpenAPI detection")
	}
	if !isMigration("db/migrations/001.sql") {
		t.Fatal("expected migration detection")
	}
	if !isDeployment("infra/main.tf") {
		t.Fatal("expected deployment detection")
	}
	if !isWorkflow(".github/workflows/ci.yml") {
		t.Fatal("expected workflow detection")
	}
}

func TestConfigDetectorAvoidsSourceFalsePositives(t *testing.T) {
	for _, p := range []string{
		"internal/servicemap/config.go",
		"internal/servicemap/config_test.go",
		"src/config.ts",
		"pkg/config.py",
	} {
		if isConfig(p) {
			t.Fatalf("isConfig(%q)=true, want false for source file", p)
		}
	}

	for _, p := range []string{
		".impactctl.yml",
		".env",
		".env.production",
		"config/app.toml",
		"config/runtime.ini",
		"deploy/app-config.yaml",
		"settings.json",
		"application.yml",
	} {
		if !isConfig(p) {
			t.Fatalf("isConfig(%q)=false, want true", p)
		}
	}
}

func TestDiscoverCodeownersStandardLocations(t *testing.T) {
	root := t.TempDir()

	docs := filepath.Join(root, "docs", "CODEOWNERS")
	if err := os.MkdirAll(filepath.Dir(docs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(docs, []byte("* @docs-team\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, ok := discoverCodeowners(root); !ok || got != docs {
		t.Fatalf("discoverCodeowners()=%q,%v want %q,true", got, ok, docs)
	}

	rootFile := filepath.Join(root, "CODEOWNERS")
	if err := os.WriteFile(rootFile, []byte("* @root-team\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, ok := discoverCodeowners(root); !ok || got != rootFile {
		t.Fatalf("root CODEOWNERS should take precedence over docs: got %q,%v", got, ok)
	}

	githubFile := filepath.Join(root, ".github", "CODEOWNERS")
	if err := os.MkdirAll(filepath.Dir(githubFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(githubFile, []byte("* @github-team\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, ok := discoverCodeowners(root); !ok || got != githubFile {
		t.Fatalf(".github/CODEOWNERS should have highest precedence: got %q,%v", got, ok)
	}
}

func TestDiscoverCodeownersMissingIsNonFatal(t *testing.T) {
	if got, ok := discoverCodeowners(t.TempDir()); ok || got != "" {
		t.Fatalf("expected missing CODEOWNERS to return empty,false; got %q,%v", got, ok)
	}
}
