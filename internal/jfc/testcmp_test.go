package jfc

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/yuin/goldmark"
)

func assertStringEqual(t testing.TB, want string, got string) {
	t.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected output (-want +got):\n%s", diff)
	}
}

func assertJSONSemanticallyEqual(t testing.TB, want []byte, got []byte) {
	t.Helper()

	var wantValue any
	wantDecoder := json.NewDecoder(bytes.NewReader(want))
	wantDecoder.UseNumber()
	if err := wantDecoder.Decode(&wantValue); err != nil {
		t.Fatalf("parse expected JSON: %v", err)
	}

	var gotValue any
	gotDecoder := json.NewDecoder(bytes.NewReader(got))
	gotDecoder.UseNumber()
	if err := gotDecoder.Decode(&gotValue); err != nil {
		t.Fatalf("parse actual JSON: %v", err)
	}

	if diff := cmp.Diff(wantValue, gotValue); diff != "" {
		t.Fatalf("JSON semantic mismatch (-want +got):\n%s", diff)
	}
}

func assertTOMLSemanticallyEqual(t testing.TB, want []byte, got []byte) {
	t.Helper()

	var wantValue any
	if err := toml.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("parse expected TOML: %v", err)
	}

	var gotValue any
	if err := toml.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("parse actual TOML: %v", err)
	}

	if diff := cmp.Diff(wantValue, gotValue); diff != "" {
		t.Fatalf("TOML semantic mismatch (-want +got):\n%s", diff)
	}
}

func assertMarkdownHTMLSemanticallyEqual(t testing.TB, want []byte, got []byte) {
	t.Helper()

	wantHTML := renderMarkdownHTML(t, want)
	gotHTML := renderMarkdownHTML(t, got)

	if diff := cmp.Diff(wantHTML, gotHTML); diff != "" {
		t.Fatalf("Markdown HTML mismatch (-want +got):\n%s", diff)
	}
}

func renderMarkdownHTML(t testing.TB, input []byte) string {
	t.Helper()

	var output bytes.Buffer
	if err := goldmark.Convert(input, &output); err != nil {
		t.Fatalf("render Markdown HTML: %v", err)
	}
	return output.String()
}
