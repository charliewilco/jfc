package jfc

import (
	"bytes"
	"testing"
)

func TestFormatJSONUsesCRLFWithoutTrailingNewline(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.EndOfLine = EndOfLineCRLF
	cfg.TrailingNewline = false
	cfg.ObjectExpand = ExpandAlways

	output, err := formatJSON([]byte(`{"x":1}`), cfg)
	if err != nil {
		t.Fatalf("formatJSON returned error: %v", err)
	}

	expected := []byte("{\r\n  \"x\": 1\r\n}")
	if !bytes.Equal(output, expected) {
		t.Fatalf("unexpected output %q", output)
	}
}

func TestFormatJSONRejectsInvalidUTF8(t *testing.T) {
	t.Parallel()

	_, err := formatJSON([]byte{0xff, 0xfe, 0xfd}, DefaultConfig())
	if err == nil || err.Error() != "input is not valid UTF-8" {
		t.Fatalf("expected UTF-8 error, got %v", err)
	}
}

func TestFormatJSONRemovesSpaceAfterColonWhenConfigured(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.ObjectExpand = ExpandNever
	cfg.SpaceAfterColon = false

	output, err := formatJSON([]byte(`{"x":1}`), cfg)
	if err != nil {
		t.Fatalf("formatJSON returned error: %v", err)
	}

	if got := string(output); got != "{\"x\":1}\n" {
		t.Fatalf("unexpected output %q", got)
	}
}

func TestFormatJSONPreservesObjectOrderByDefault(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.ObjectExpand = ExpandNever

	output, err := formatJSON([]byte(`{"z":1,"a":2}`), cfg)
	if err != nil {
		t.Fatalf("formatJSON returned error: %v", err)
	}

	if got := string(output); got != "{\"z\": 1, \"a\": 2}\n" {
		t.Fatalf("unexpected output %q", got)
	}
}

func TestFormatJSONAppliesSpacingToEmptyContainers(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.SpaceWithinBraces = true
	cfg.SpaceWithinBrackets = true
	cfg.ObjectExpand = ExpandNever
	cfg.ArrayExpand = ExpandNever

	output, err := formatJSON([]byte(`{"emptyObject":{},"emptyArray":[]}`), cfg)
	if err != nil {
		t.Fatalf("formatJSON returned error: %v", err)
	}

	if got := string(output); got != "{ \"emptyObject\": { }, \"emptyArray\": [ ] }\n" {
		t.Fatalf("unexpected output %q", got)
	}
}

func TestFormatJSONAutoExpandsWhenPrintWidthIsTooSmall(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.PrintWidth = 10

	output, err := formatJSON([]byte(`{"alpha":[1,2,3]}`), cfg)
	if err != nil {
		t.Fatalf("formatJSON returned error: %v", err)
	}

	expected := "{\n  \"alpha\": [\n    1,\n    2,\n    3\n  ]\n}\n"
	if string(output) != expected {
		t.Fatalf("unexpected output:\n%s", output)
	}
}
