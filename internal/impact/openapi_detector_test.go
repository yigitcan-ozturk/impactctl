package impact

import "testing"

func TestOpenAPIDetectorRequiresContractDocumentType(t *testing.T) {
	for _, path := range []string{
		"internal/servicemap/openapi.go",
		"internal/servicemap/openapi_test.go",
		"src/openapi.ts",
		"docs/swagger.md",
	} {
		if isOpenAPI(path) {
			t.Fatalf("isOpenAPI(%q)=true, want false for non-contract source/document", path)
		}
	}

	for _, path := range []string{
		"api/openapi.yaml",
		"api/openapi.json",
		"spec/swagger.yml",
		"api/api.yaml",
		"api/api.yml",
		"api/api.json",
	} {
		if !isOpenAPI(path) {
			t.Fatalf("isOpenAPI(%q)=false, want true", path)
		}
	}
}
