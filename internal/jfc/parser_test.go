package jfc

import (
	"strings"
	"testing"
)

func TestParseJSONParsesEscapedStringsAndNumbers(t *testing.T) {
	t.Parallel()

	root, err := parseJSON([]byte(`{"message":"hello\nworld","value":1.25e+3}`))
	if err != nil {
		t.Fatalf("parseJSON returned error: %v", err)
	}
	if root.kind != kindObject {
		t.Fatalf("expected object root, got %v", root.kind)
	}
	if len(root.object) != 2 {
		t.Fatalf("expected 2 object members, got %d", len(root.object))
	}
	if root.object[0].key != "message" || root.object[0].value.raw != "hello\nworld" {
		t.Fatalf("unexpected parsed string member: %+v", root.object[0])
	}
	if root.object[1].key != "value" || root.object[1].value.raw != "1.25e+3" {
		t.Fatalf("unexpected parsed number member: %+v", root.object[1])
	}
}

func TestParseJSONRejectsTrailingContent(t *testing.T) {
	t.Parallel()

	_, err := parseJSON([]byte(`{} true`))
	if err == nil || !strings.Contains(err.Error(), "unexpected trailing content") {
		t.Fatalf("expected trailing content error, got %v", err)
	}
}

func TestParseJSONRejectsInvalidNumber(t *testing.T) {
	t.Parallel()

	_, err := parseJSON([]byte(`{"n": 01}`))
	if err == nil || !strings.Contains(err.Error(), "leading zeroes are not allowed") {
		t.Fatalf("expected invalid number error, got %v", err)
	}
}

func TestParseJSONRejectsInvalidUnicodeEscape(t *testing.T) {
	t.Parallel()

	_, err := parseJSON([]byte(`{"bad":"\u12xz"}`))
	if err == nil || !strings.Contains(err.Error(), "invalid unicode escape") {
		t.Fatalf("expected invalid unicode escape error, got %v", err)
	}
}
