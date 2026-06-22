package jfc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatDocumentFixtures(t *testing.T) {
	t.Parallel()

	fixtures := []struct {
		name   string
		input  string
		golden string
	}{
		{name: "json", input: "json.input.json", golden: "json.golden.json"},
		{name: "package_json", input: "package.input.json", golden: "package.golden.json"},
		{name: "tsconfig_json", input: "tsconfig.input.json", golden: "tsconfig.golden.json"},
		{name: "jsonc", input: "jsonc.input.jsonc", golden: "jsonc.golden.jsonc"},
		{name: "settings_jsonc", input: "settings.input.jsonc", golden: "settings.golden.jsonc"},
		{name: "jsonl", input: "jsonl.input.jsonl", golden: "jsonl.golden.jsonl"},
		{name: "events_ndjson", input: "events.input.ndjson", golden: "events.golden.ndjson"},
		{name: "yaml", input: "yaml.input.yaml", golden: "yaml.golden.yaml"},
		{name: "workflow_yaml", input: "workflow.input.yaml", golden: "workflow.golden.yaml"},
		{name: "docker_compose_yaml", input: "docker_compose.input.yaml", golden: "docker_compose.golden.yaml"},
		{name: "toml", input: "toml.input.toml", golden: "toml.golden.toml"},
		{name: "toml_multiline", input: "toml_multiline.input.toml", golden: "toml_multiline.golden.toml"},
		{name: "toml_edges", input: "toml_edges.input.toml", golden: "toml_edges.golden.toml"},
		{name: "pyproject_toml", input: "pyproject.input.toml", golden: "pyproject.golden.toml"},
		{name: "cargo_toml", input: "cargo.input.toml", golden: "cargo.golden.toml"},
		{name: "markdown", input: "markdown.input.md", golden: "markdown.golden.md"},
		{name: "markdown_fences", input: "markdown_fences.input.md", golden: "markdown_fences.golden.md"},
		{name: "readme_markdown", input: "readme.input.md", golden: "readme.golden.md"},
		{name: "changelog_markdown", input: "changelog.input.md", golden: "changelog.golden.md"},
		{name: "xml", input: "xml.input.xml", golden: "xml.golden.xml"},
		{name: "svg", input: "svg.input.svg", golden: "svg.golden.svg"},
		{name: "plist", input: "plist.input.plist", golden: "plist.golden.plist"},
		{name: "csproj", input: "csproj.input.csproj", golden: "csproj.golden.csproj"},
		{name: "csv", input: "csv.input.csv", golden: "csv.golden.csv"},
		{name: "tsv", input: "tsv.input.tsv", golden: "tsv.golden.tsv"},
		{name: "dotenv", input: "dotenv.input.env", golden: "dotenv.golden.env"},
		{name: "terraform", input: "terraform.input.tf", golden: "terraform.golden.tf"},
	}

	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()

			inputPath := filepath.Join("testdata", "format", fixture.input)
			goldenPath := filepath.Join("testdata", "format", fixture.golden)

			input, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("read input fixture: %v", err)
			}
			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden fixture: %v", err)
			}

			format, ok := detectFormat(inputPath)
			if !ok {
				t.Fatalf("fixture input has unsupported extension: %s", inputPath)
			}

			output, err := formatDocument(input, format, DefaultConfig())
			if err != nil {
				t.Fatalf("formatDocument returned error: %v", err)
			}
			assertStringEqual(t, string(expected), string(output))
			assertFixtureSemanticsEqual(t, format, input, output)

			idempotent, err := formatDocument(expected, format, DefaultConfig())
			if err != nil {
				t.Fatalf("formatDocument returned error for golden output: %v", err)
			}
			assertStringEqual(t, string(expected), string(idempotent))
		})
	}
}

func TestFormatJSONCSortCommentFixture(t *testing.T) {
	t.Parallel()

	inputPath := filepath.Join("testdata", "format", "jsonc_sort_comments.input.jsonc")
	goldenPath := filepath.Join("testdata", "format", "jsonc_sort_comments.golden.jsonc")

	input, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("read input fixture: %v", err)
	}
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden fixture: %v", err)
	}

	cfg := DefaultConfig()
	cfg.SortKeys = true
	output, err := formatJSONC(input, cfg)
	if err != nil {
		t.Fatalf("formatJSONC returned error: %v", err)
	}
	assertStringEqual(t, string(expected), string(output))

	idempotent, err := formatJSONC(expected, cfg)
	if err != nil {
		t.Fatalf("formatJSONC returned error for golden output: %v", err)
	}
	assertStringEqual(t, string(expected), string(idempotent))
}

func assertFixtureSemanticsEqual(t testing.TB, format FormatKind, input []byte, output []byte) {
	t.Helper()

	switch format {
	case FormatJSON:
		assertJSONSemanticallyEqual(t, input, output)
	case FormatTOML:
		assertTOMLSemanticallyEqual(t, input, output)
	case FormatYAML:
		assertYAMLStreamSemanticallyEqual(t, input, output)
	case FormatMarkdown:
		assertMarkdownHTMLSemanticallyEqual(t, input, output)
	}
}
