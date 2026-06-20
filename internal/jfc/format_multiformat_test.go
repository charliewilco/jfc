package jfc

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatJSONCPreservesCommentsAndAcceptsTrailingCommas(t *testing.T) {
	t.Parallel()

	input := []byte("{\n// keep this\n\"b\": 1,\n\"a\": [1,2,],\n}\n")
	output, err := formatJSONC(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatJSONC returned error: %v", err)
	}

	got := string(output)
	if !strings.Contains(got, "// keep this") {
		t.Fatalf("expected comment to be preserved, got:\n%s", got)
	}
	if !strings.Contains(got, `"a"`) || !strings.Contains(got, `"b"`) {
		t.Fatalf("expected object keys to be preserved, got:\n%s", got)
	}
}

func TestFormatJSONCSortsKeysWhenEnabled(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.SortKeys = true
	output, err := formatJSONC([]byte("{\n\"z\": 1,\n// keep with a\n\"a\": {\"b\": 2, \"a\": 1},\n}\n"), cfg)
	if err != nil {
		t.Fatalf("formatJSONC returned error: %v", err)
	}

	got := string(output)
	aIndex := strings.Index(got, `"a"`)
	zIndex := strings.Index(got, `"z"`)
	if aIndex < 0 || zIndex < 0 || aIndex > zIndex {
		t.Fatalf("expected sorted JSONC keys, got:\n%s", got)
	}
	if !strings.Contains(got, "// keep with a") {
		t.Fatalf("expected comment to be preserved, got:\n%s", got)
	}
}

func TestFormatJSONLFormatsEachRecordInline(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.SortKeys = true
	output, err := formatJSONL([]byte("{\"z\":1,\"a\":2}\n\n{\"items\":[1,2]}\n"), cfg)
	if err != nil {
		t.Fatalf("formatJSONL returned error: %v", err)
	}

	expected := "{\"a\": 2, \"z\": 1}\n{\"items\": [1, 2]}\n"
	assertStringEqual(t, expected, string(output))
}

func TestFormatJSONLReportsLineSpecificErrors(t *testing.T) {
	t.Parallel()

	_, err := formatJSONL([]byte("{\"ok\":true}\n{\"bad\":}\n"), DefaultConfig())
	if err == nil || !strings.Contains(err.Error(), "line 2:") {
		t.Fatalf("expected line-specific JSONL error, got %v", err)
	}
}

func TestFormatYAMLPreservesCommentsAndUsesSpaceIndent(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.UseTabs = true
	cfg.TabWidth = 4
	output, err := formatYAML([]byte("root:\n  # keep this\n  child: value\n"), cfg)
	if err != nil {
		t.Fatalf("formatYAML returned error: %v", err)
	}

	got := string(output)
	if !strings.Contains(got, "# keep this") {
		t.Fatalf("expected YAML comment to be preserved, got:\n%s", got)
	}
	if bytes.Contains(output, []byte("\t")) {
		t.Fatalf("expected YAML indentation to use spaces, got:\n%s", output)
	}
}

func TestFormatTOMLValidatesAndNormalizesAssignments(t *testing.T) {
	t.Parallel()

	input := []byte("title=\"jfc\" # keep this\n\n[tool]\nitems=[\"a\", \"b\"]\n")
	output, err := formatTOML(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatTOML returned error: %v", err)
	}

	expected := "title = \"jfc\" # keep this\n\n[tool]\nitems = [\"a\", \"b\"]\n"
	assertStringEqual(t, expected, string(output))
}

func TestFormatTOMLPreservesEqualsInsideStringsAndComments(t *testing.T) {
	t.Parallel()

	input := []byte(strings.Join([]string{
		`basic="a=b" # comment has c=d`,
		`literal='x=y'`,
		`escaped="quote \" = still inside"`,
		`url="https://example.test/search?q=a=b"`,
		`# comment_only=unchanged`,
		`# not_assignment # still not key=value`,
		``,
	}, "\n"))
	output, err := formatTOML(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatTOML returned error: %v", err)
	}

	expected := strings.Join([]string{
		`basic = "a=b" # comment has c=d`,
		`literal = 'x=y'`,
		`escaped = "quote \" = still inside"`,
		`url = "https://example.test/search?q=a=b"`,
		`# comment_only=unchanged`,
		`# not_assignment # still not key=value`,
		``,
	}, "\n")
	assertStringEqual(t, expected, string(output))
}

func TestFormatTOMLPreservesMultilineStringBodies(t *testing.T) {
	t.Parallel()

	input := []byte(strings.Join([]string{
		`basic="""`,
		`alpha=beta  `,
		`# not a comment`,
		`"""  `,
		`literal='''`,
		`gamma=delta	`,
		`'''`,
		`after="done"`,
		``,
	}, "\n"))
	output, err := formatTOML(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatTOML returned error: %v", err)
	}

	expected := strings.Join([]string{
		`basic = """`,
		`alpha=beta  `,
		`# not a comment`,
		`"""  `,
		`literal = '''`,
		`gamma=delta	`,
		`'''`,
		`after = "done"`,
		``,
	}, "\n")
	assertStringEqual(t, expected, string(output))
	assertTOMLSemanticallyEqual(t, input, output)
}

func TestFormatTOMLPreservesSameLineMultilineStringContent(t *testing.T) {
	t.Parallel()

	input := []byte(strings.Join([]string{
		`basic="""alpha=beta  `,
		`"""`,
		`literal='''gamma=delta	`,
		`'''`,
		``,
	}, "\n"))
	output, err := formatTOML(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatTOML returned error: %v", err)
	}

	expected := strings.Join([]string{
		`basic = """alpha=beta  `,
		`"""`,
		`literal = '''gamma=delta	`,
		`'''`,
		``,
	}, "\n")
	assertStringEqual(t, expected, string(output))
	assertTOMLSemanticallyEqual(t, input, output)
}

func TestFormatTOMLDoesNotNormalizeEqualsInsideArrayValues(t *testing.T) {
	t.Parallel()

	input := []byte(strings.Join([]string{
		`items=[`,
		`  {x=1,y=2},`,
		`  {url="https://example.test?q=a=b"},`,
		`]`,
		`point={x=1,y=2}`,
		``,
	}, "\n"))
	output, err := formatTOML(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatTOML returned error: %v", err)
	}

	expected := strings.Join([]string{
		`items = [`,
		`  {x=1,y=2},`,
		`  {url="https://example.test?q=a=b"},`,
		`]`,
		`point = {x=1,y=2}`,
		``,
	}, "\n")
	assertStringEqual(t, expected, string(output))
	assertTOMLSemanticallyEqual(t, input, output)
}

func TestFormatMarkdownConservativelyNormalizesWhitespace(t *testing.T) {
	t.Parallel()

	input := []byte("# Title\r\n   \r\n  ```go\r\n  fmt.Println(\"kept\")  \r\n  ```\r\n")
	output, err := formatMarkdown(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatMarkdown returned error: %v", err)
	}

	expected := "# Title\n\n  ```go\n  fmt.Println(\"kept\")  \n  ```\n"
	assertStringEqual(t, expected, string(output))
}

func TestFormatMarkdownDoesNotTreatIndentedCodeAsFence(t *testing.T) {
	t.Parallel()

	input := []byte("    ```go\n    fmt.Println(\"kept\")\n    ```\n")
	output, err := formatMarkdown(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatMarkdown returned error: %v", err)
	}

	assertStringEqual(t, string(input), string(output))
}

func TestFormatMarkdownRequiresMatchingFenceClose(t *testing.T) {
	t.Parallel()

	input := []byte(strings.Join([]string{
		"  ````markdown",
		"  ```",
		"   ",
		"  ````",
		"",
	}, "\n"))
	output, err := formatMarkdown(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatMarkdown returned error: %v", err)
	}

	assertStringEqual(t, string(input), string(output))
	assertMarkdownHTMLSemanticallyEqual(t, input, output)
}

func TestFormatMarkdownDoesNotCloseFenceWithInfoText(t *testing.T) {
	t.Parallel()

	input := []byte(strings.Join([]string{
		"```go",
		"``` still code",
		"   ",
		"```",
		"",
	}, "\n"))
	output, err := formatMarkdown(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatMarkdown returned error: %v", err)
	}

	assertStringEqual(t, string(input), string(output))
	assertMarkdownHTMLSemanticallyEqual(t, input, output)
}

func TestFormatMarkdownDoesNotOpenBacktickFenceWithBacktickInfo(t *testing.T) {
	t.Parallel()

	input := []byte("``` `not an opener`\nblank follows   \n\n")
	output, err := formatMarkdown(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatMarkdown returned error: %v", err)
	}

	expected := "``` `not an opener`\nblank follows   \n"
	assertStringEqual(t, expected, string(output))
	assertMarkdownHTMLSemanticallyEqual(t, input, output)
}
