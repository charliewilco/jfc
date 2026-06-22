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

func TestFormatYAMLPreservesMultiDocumentStreams(t *testing.T) {
	t.Parallel()

	input := []byte("---\na: 1\n---\nb: 2\n")
	output, err := formatYAML(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatYAML returned error: %v", err)
	}

	expected := "a: 1\n---\nb: 2\n"
	assertStringEqual(t, expected, string(output))
	assertYAMLStreamSemanticallyEqual(t, input, output)
}

func TestFormatYAMLPreservesFeatureRichDocuments(t *testing.T) {
	t.Parallel()

	input := []byte(strings.Join([]string{
		"%YAML 1.1",
		"---",
		"# keep leading comment",
		"defaults: &defaults",
		"  image: !!str 1.0",
		"  command: >",
		"    echo alpha",
		"    echo beta",
		"service:",
		"  <<: *defaults",
		"  tag: !Ref value",
		"  list:",
		"    - first",
		"    - second # keep inline comment",
		"---",
		"ordered:",
		"  z: 1",
		"  a: 2",
		"",
	}, "\n"))
	output, err := formatYAML(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatYAML returned error: %v", err)
	}

	got := string(output)
	for _, fragment := range []string{
		"# keep leading comment",
		"&defaults",
		"*defaults",
		"!!str 1.0",
		"!Ref value",
		">",
		"# keep inline comment",
		"z: 1\n  a: 2",
	} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected YAML output to contain %q, got:\n%s", fragment, got)
		}
	}
	assertYAMLStreamSemanticallyEqual(t, input, output)
}

func TestFormatYAMLEmptyInputFormatsAsNull(t *testing.T) {
	t.Parallel()

	output, err := formatYAML(nil, DefaultConfig())
	if err != nil {
		t.Fatalf("formatYAML returned error: %v", err)
	}

	assertStringEqual(t, "null\n", string(output))
}

func TestFormatXMLPrettyPrintsElementOnlyDocuments(t *testing.T) {
	t.Parallel()

	input := []byte(`<?xml version="1.0"?><!-- keep --><root b="2" a="1"><child/><nested><leaf name="ok"/></nested></root>`)
	output, err := formatXML(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatXML returned error: %v", err)
	}

	expected := strings.Join([]string{
		`<?xml version="1.0"?>`,
		`<!-- keep -->`,
		`<root b="2" a="1">`,
		`  <child/>`,
		`  <nested>`,
		`    <leaf name="ok"/>`,
		`  </nested>`,
		`</root>`,
		``,
	}, "\n")
	assertStringEqual(t, expected, string(output))
}

func TestFormatXMLPreservesDirectivesAndProcessingInstructions(t *testing.T) {
	t.Parallel()

	input := []byte(`<?xml version="1.0"?><!DOCTYPE note SYSTEM "Note.dtd"><?xml-stylesheet href="style.css"?><note><to>Tove</to></note>`)
	output, err := formatXML(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatXML returned error: %v", err)
	}

	expected := strings.Join([]string{
		`<?xml version="1.0"?>`,
		`<!DOCTYPE note SYSTEM "Note.dtd">`,
		`<?xml-stylesheet href="style.css"?>`,
		`<note>`,
		`  <to>Tove</to>`,
		`</note>`,
		``,
	}, "\n")
	assertStringEqual(t, expected, string(output))
}

func TestFormatXMLUsesTabsWhenConfigured(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.UseTabs = true
	output, err := formatXML([]byte(`<root><child><leaf/></child></root>`), cfg)
	if err != nil {
		t.Fatalf("formatXML returned error: %v", err)
	}

	expected := "<root>\n\t<child>\n\t\t<leaf/>\n\t</child>\n</root>\n"
	assertStringEqual(t, expected, string(output))
}

func TestFormatXMLRejectsMalformedInput(t *testing.T) {
	t.Parallel()

	_, err := formatXML([]byte(`<root><child></root>`), DefaultConfig())
	if err == nil || !strings.Contains(err.Error(), "expected </child>") {
		t.Fatalf("expected XML mismatch error, got %v", err)
	}
}

func TestFormatXMLFallsBackForMixedTextContent(t *testing.T) {
	t.Parallel()

	input := []byte("<p>Hello <strong>world</strong>!</p>\r\n")
	output, err := formatXML(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatXML returned error: %v", err)
	}

	assertStringEqual(t, "<p>Hello <strong>world</strong>!</p>\n", string(output))
}

func TestFormatXMLFallsBackForCDATA(t *testing.T) {
	t.Parallel()

	input := []byte("<root><![CDATA[<not-xml>]]></root>")
	output, err := formatXML(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatXML returned error: %v", err)
	}

	assertStringEqual(t, "<root><![CDATA[<not-xml>]]></root>\n", string(output))
}

func TestFormatCSVValidatesAndPreservesBytes(t *testing.T) {
	t.Parallel()

	input := []byte("name,notes\r\nalice,\"hello\r\nworld\"")
	output, err := formatCSV(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatCSV returned error: %v", err)
	}

	assertStringEqual(t, "name,notes\r\nalice,\"hello\r\nworld\"\n", string(output))
}

func TestFormatCSVRejectsMalformedInput(t *testing.T) {
	t.Parallel()

	_, err := formatCSV([]byte("name,notes\nalice,\"unterminated\n"), DefaultConfig())
	if err == nil || !strings.Contains(err.Error(), "extraneous or missing") {
		t.Fatalf("expected CSV parse error, got %v", err)
	}
}

func TestFormatTSVValidatesAndPreservesBytes(t *testing.T) {
	t.Parallel()

	input := []byte("name\tnotes\nalice\thello")
	output, err := formatTSV(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatTSV returned error: %v", err)
	}

	assertStringEqual(t, "name\tnotes\nalice\thello\n", string(output))
}

func TestFormatDotenvNormalizesAssignments(t *testing.T) {
	t.Parallel()

	input := []byte("# keep\r\n export KEY = value  \nQUOTED = \" spaced value \"\nINLINE = value # comment\nNO_VALUE\n")
	output, err := formatDotenv(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatDotenv returned error: %v", err)
	}

	expected := strings.Join([]string{
		"# keep",
		"export KEY=value",
		"QUOTED=\" spaced value \"",
		"INLINE=value # comment",
		"NO_VALUE",
		"",
	}, "\n")
	assertStringEqual(t, expected, string(output))
}

func TestFormatDotenvRejectsInvalidAssignmentKey(t *testing.T) {
	t.Parallel()

	_, err := formatDotenv([]byte("1BAD=value\n"), DefaultConfig())
	if err == nil || !strings.Contains(err.Error(), "invalid dotenv key") {
		t.Fatalf("expected dotenv key error, got %v", err)
	}
}

func TestFormatHCLFormatsWithHashiCorpStyle(t *testing.T) {
	t.Parallel()

	input := []byte("resource \"thing\" \"example\" {\nname=\"ok\"\nsetting {\nvalue=1\n}\n}\n")
	output, err := formatHCL(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatHCL returned error: %v", err)
	}

	expected := strings.Join([]string{
		`resource "thing" "example" {`,
		`  name = "ok"`,
		`  setting {`,
		`    value = 1`,
		`  }`,
		`}`,
		``,
	}, "\n")
	assertStringEqual(t, expected, string(output))
}

func TestFormatHCLRejectsMalformedInput(t *testing.T) {
	t.Parallel()

	_, err := formatHCL([]byte(`resource "thing" "example" {`), DefaultConfig())
	if err == nil {
		t.Fatal("expected HCL parse error")
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

func TestFormatMarkdownPreservesIndentedCodeBlankLines(t *testing.T) {
	t.Parallel()

	input := []byte("    alpha\n    \n    beta\n\t\n")
	output, err := formatMarkdown(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatMarkdown returned error: %v", err)
	}

	assertStringEqual(t, string(input), string(output))
	assertMarkdownHTMLSemanticallyEqual(t, input, output)
}

func TestFormatMarkdownDoesNotTreatVerticalTabAsBlankLine(t *testing.T) {
	t.Parallel()

	input := []byte("\v")
	output, err := formatMarkdown(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatMarkdown returned error: %v", err)
	}

	assertStringEqual(t, string(input)+"\n", string(output))
	assertMarkdownHTMLSemanticallyEqual(t, input, output)
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

func TestFormatMarkdownDoesNotAddNewlineToUnclosedFence(t *testing.T) {
	t.Parallel()

	input := []byte("```0")
	output, err := formatMarkdown(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatMarkdown returned error: %v", err)
	}

	assertStringEqual(t, string(input), string(output))
	assertMarkdownHTMLSemanticallyEqual(t, input, output)
}

func TestFormatMarkdownPreservesTrailingNewlinesInUnclosedFence(t *testing.T) {
	t.Parallel()

	input := []byte("```\n\n")
	output, err := formatMarkdown(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatMarkdown returned error: %v", err)
	}

	assertStringEqual(t, string(input), string(output))
	assertMarkdownHTMLSemanticallyEqual(t, input, output)
}
