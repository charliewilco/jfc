package jfc

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/yuin/goldmark"
	"gopkg.in/yaml.v3"
)

func assertStringEqual(t testing.TB, want string, got string) {
	t.Helper()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected output (-want +got):\n%s", diff)
	}
}

func assertJSONSemanticallyEqual(t testing.TB, want []byte, got []byte) {
	t.Helper()

	wantValue, err := decodeStrictJSON(want)
	if err != nil {
		t.Fatalf("parse expected JSON: %v", err)
	}

	gotValue, err := decodeStrictJSON(got)
	if err != nil {
		t.Fatalf("parse actual JSON: %v", err)
	}

	if diff := cmp.Diff(wantValue, gotValue); diff != "" {
		t.Fatalf("JSON semantic mismatch (-want +got):\n%s", diff)
	}
}

func decodeStrictJSON(input []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(input))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	return value, nil
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

	if diff := cmp.Diff(wantValue, gotValue, cmpopts.EquateNaNs()); diff != "" {
		t.Fatalf("TOML semantic mismatch (-want +got):\n%s", diff)
	}
}

func assertYAMLStreamSemanticallyEqual(t testing.TB, want []byte, got []byte) {
	t.Helper()

	wantDocuments, err := decodeYAMLStream(want)
	if err != nil {
		t.Fatalf("parse expected YAML stream: %v", err)
	}

	gotDocuments, err := decodeYAMLStream(got)
	if err != nil {
		t.Fatalf("parse actual YAML stream: %v", err)
	}

	if diff := cmp.Diff(wantDocuments, gotDocuments, cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("YAML stream semantic mismatch (-want +got):\n%s", diff)
	}
}

func decodeYAMLStream(input []byte) ([]any, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(input))
	var documents []any
	for {
		var document any
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}
	if len(documents) == 0 {
		documents = append(documents, nil)
	}
	return documents, nil
}

func assertMarkdownHTMLSemanticallyEqual(t testing.TB, want []byte, got []byte) {
	t.Helper()

	wantHTML := renderMarkdownHTML(t, []byte(normalizeLineEndingsToLF(string(want))))
	gotHTML := renderMarkdownHTML(t, []byte(normalizeLineEndingsToLF(string(got))))

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
