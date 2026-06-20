package jfc

import (
	"strings"
	"testing"
)

func BenchmarkFormatJSONProjectConfig(b *testing.B) {
	input := []byte(`{"scripts":{"test":"go test ./...","build":"go build ./..."},"dependencies":{"goldmark":"1.8.2","hujson":"latest"},"nested":[{"name":"alpha","enabled":true},{"name":"beta","enabled":false}],"numbers":[1,2,3,4,5]}`)
	cfg := DefaultConfig()

	b.ReportAllocs()
	b.SetBytes(int64(len(input)))
	for b.Loop() {
		if _, err := formatJSON(input, cfg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFormatJSONLargeArray(b *testing.B) {
	input := []byte("[" + strings.Repeat(`{"id":123,"name":"example","ok":true},`, 999) + `{"id":123,"name":"example","ok":true}]`)
	cfg := DefaultConfig()

	b.ReportAllocs()
	b.SetBytes(int64(len(input)))
	for b.Loop() {
		if _, err := formatJSON(input, cfg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFormatTOMLProjectConfig(b *testing.B) {
	input := []byte(strings.Join([]string{
		`[project]`,
		`name="jfc"`,
		`version="1.0.0"`,
		`requires-python=">=3.12"`,
		``,
		`[tool.jfc]`,
		`sort_keys=true`,
		`ignore=["dist", "*.generated.json"]`,
		``,
		`[[tool.jfc.fixture]]`,
		`name="package"`,
		`path="package.json"`,
		``,
	}, "\n"))
	cfg := DefaultConfig()

	b.ReportAllocs()
	b.SetBytes(int64(len(input)))
	for b.Loop() {
		if _, err := formatTOML(input, cfg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFormatMarkdownReadme(b *testing.B) {
	input := []byte(strings.Repeat("# Title\r\n   \r\nParagraph with `code`.\r\n\r\n```go\r\nfmt.Println(\"x\")\r\n```\r\n", 100))
	cfg := DefaultConfig()

	b.ReportAllocs()
	b.SetBytes(int64(len(input)))
	for b.Loop() {
		if _, err := formatMarkdown(input, cfg); err != nil {
			b.Fatal(err)
		}
	}
}
