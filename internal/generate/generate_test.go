package generate

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func TestRenderMigrationTemplatesAgainstGolden(t *testing.T) {
	t.Parallel()

	templateData := migrationTemplateData{
		Name:      "create_products",
		Timestamp: "2026-01-01T00:00:00Z",
		TableName: "products",
	}

	upTemplateBytes, readUpTemplateError := templatesFS.ReadFile("templates/migration_up.sql.tmpl")
	if readUpTemplateError != nil {
		t.Fatalf("read up template: %v", readUpTemplateError)
	}
	upOutputBytes, renderUpTemplateError := RenderTemplate(string(upTemplateBytes), templateData, "migration_up.sql.tmpl")
	if renderUpTemplateError != nil {
		t.Fatalf("render up template: %v", renderUpTemplateError)
	}
	assertGolden(t, "testdata/product_migration_up.sql.golden", upOutputBytes)

	downTemplateBytes, readDownTemplateError := templatesFS.ReadFile("templates/migration_down.sql.tmpl")
	if readDownTemplateError != nil {
		t.Fatalf("read down template: %v", readDownTemplateError)
	}
	downOutputBytes, renderDownTemplateError := RenderTemplate(string(downTemplateBytes), templateData, "migration_down.sql.tmpl")
	if renderDownTemplateError != nil {
		t.Fatalf("render down template: %v", renderDownTemplateError)
	}
	assertGolden(t, "testdata/product_migration_down.sql.golden", downOutputBytes)
}

func TestRenderResourceTemplatesAgainstGolden(t *testing.T) {
	t.Parallel()

	resourceData := ResourceData{
		ModulePath:      "github.com/eminbekov/fiber-v3-template",
		EntityName:      "Product",
		EntityNameLower: "product",
		EntityVariable:  "product",
		TableName:       "products",
		FileName:        "product",
		APIVersion:      "v1",
	}

	domainTemplateBytes, readDomainTemplateError := templatesFS.ReadFile("templates/domain.go.tmpl")
	if readDomainTemplateError != nil {
		t.Fatalf("read domain template: %v", readDomainTemplateError)
	}
	domainOutputBytes, renderDomainTemplateError := RenderTemplate(string(domainTemplateBytes), resourceData, "domain.go.tmpl")
	if renderDomainTemplateError != nil {
		t.Fatalf("render domain template: %v", renderDomainTemplateError)
	}
	assertGolden(t, "testdata/product_domain.go.golden", domainOutputBytes)

	requestTemplateBytes, readRequestTemplateError := templatesFS.ReadFile("templates/dto_request.go.tmpl")
	if readRequestTemplateError != nil {
		t.Fatalf("read request template: %v", readRequestTemplateError)
	}
	requestOutputBytes, renderRequestTemplateError := RenderTemplate(string(requestTemplateBytes), resourceData, "dto_request.go.tmpl")
	if renderRequestTemplateError != nil {
		t.Fatalf("render request template: %v", renderRequestTemplateError)
	}
	assertGolden(t, "testdata/product_dto_request.go.golden", requestOutputBytes)
}

func TestWriteFileDryRunNoWrite(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "sample.txt")
	if writeFileError := WriteFile(outputPath, []byte("sample"), true, false); writeFileError != nil {
		t.Fatalf("write file dry-run: %v", writeFileError)
	}
	if _, statError := os.Stat(outputPath); !os.IsNotExist(statError) {
		t.Fatalf("expected file to not exist after dry-run")
	}
}

func TestWriteFileNoOverwriteWithoutForce(t *testing.T) {
	t.Parallel()

	outputPath := filepath.Join(t.TempDir(), "sample.txt")
	if writeError := os.WriteFile(outputPath, []byte("original"), 0o600); writeError != nil {
		t.Fatalf("seed file: %v", writeError)
	}
	writeFileError := WriteFile(outputPath, []byte("new"), false, false)
	if writeFileError == nil || !strings.Contains(writeFileError.Error(), "output already exists") {
		t.Fatalf("expected overwrite protection error, got: %v", writeFileError)
	}
}

func TestNamingConversions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		input           string
		expectedSnake   string
		expectedCamel   string
		expectedLower   string
		expectedPlural  string
		expectedTable   string
	}{
		{
			name:           "camel case name",
			input:          "OrderItem",
			expectedSnake:  "order_item",
			expectedCamel:  "OrderItem",
			expectedLower:  "orderItem",
			expectedPlural: "order_items",
			expectedTable:  "order_items",
		},
		{
			name:           "snake case name",
			input:          "category",
			expectedSnake:  "category",
			expectedCamel:  "Category",
			expectedLower:  "category",
			expectedPlural: "categories",
			expectedTable:  "categories",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if actual := ToSnakeCase(testCase.input); actual != testCase.expectedSnake {
				t.Fatalf("ToSnakeCase mismatch: got %q want %q", actual, testCase.expectedSnake)
			}
			if actual := ToCamelCase(testCase.input); actual != testCase.expectedCamel {
				t.Fatalf("ToCamelCase mismatch: got %q want %q", actual, testCase.expectedCamel)
			}
			if actual := ToLowerCamelCase(testCase.input); actual != testCase.expectedLower {
				t.Fatalf("ToLowerCamelCase mismatch: got %q want %q", actual, testCase.expectedLower)
			}
			if actual := Pluralize(ToSnakeCase(testCase.input)); actual != testCase.expectedPlural {
				t.Fatalf("Pluralize mismatch: got %q want %q", actual, testCase.expectedPlural)
			}
			if actual := TableName(testCase.input); actual != testCase.expectedTable {
				t.Fatalf("TableName mismatch: got %q want %q", actual, testCase.expectedTable)
			}
		})
	}
}

func assertGolden(t *testing.T, relativeGoldenPath string, outputBytes []byte) {
	t.Helper()

	goldenPath := relativeGoldenPath
	if *updateGolden {
		if writeGoldenError := os.WriteFile(goldenPath, outputBytes, 0o600); writeGoldenError != nil {
			t.Fatalf("update golden %s: %v", goldenPath, writeGoldenError)
		}
	}

	//nolint:gosec // test loads repo-owned golden files.
	expectedBytes, readGoldenError := os.ReadFile(goldenPath)
	if readGoldenError != nil {
		t.Fatalf("read golden %s: %v", goldenPath, readGoldenError)
	}
	if string(expectedBytes) != string(outputBytes) {
		t.Fatalf("golden mismatch for %s\nexpected:\n%s\nactual:\n%s", relativeGoldenPath, string(expectedBytes), string(outputBytes))
	}
}
