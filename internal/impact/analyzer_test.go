package impact

import "testing"

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
