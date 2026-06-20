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
		{name: "toml", input: "toml.input.toml", golden: "toml.golden.toml"},
		{name: "pyproject_toml", input: "pyproject.input.toml", golden: "pyproject.golden.toml"},
		{name: "markdown", input: "markdown.input.md", golden: "markdown.golden.md"},
		{name: "markdown_fences", input: "markdown_fences.input.md", golden: "markdown_fences.golden.md"},
		{name: "readme_markdown", input: "readme.input.md", golden: "readme.golden.md"},
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
			if string(output) != string(expected) {
				t.Fatalf("fixture output mismatch\nexpected:\n%s\nactual:\n%s", expected, output)
			}

			idempotent, err := formatDocument(expected, format, DefaultConfig())
			if err != nil {
				t.Fatalf("formatDocument returned error for golden output: %v", err)
			}
			if string(idempotent) != string(expected) {
				t.Fatalf("fixture golden output is not idempotent\nexpected:\n%s\nactual:\n%s", expected, idempotent)
			}
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
	if string(output) != string(expected) {
		t.Fatalf("fixture output mismatch\nexpected:\n%s\nactual:\n%s", expected, output)
	}

	idempotent, err := formatJSONC(expected, cfg)
	if err != nil {
		t.Fatalf("formatJSONC returned error for golden output: %v", err)
	}
	if string(idempotent) != string(expected) {
		t.Fatalf("fixture golden output is not idempotent\nexpected:\n%s\nactual:\n%s", expected, idempotent)
	}
}
