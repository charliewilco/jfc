package jfc

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	toml "github.com/pelletier/go-toml/v2"
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
