package jfc

import "testing"

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
	assertStringEqual(t, string(expected), string(output))
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

	assertStringEqual(t, "{\"x\":1}\n", string(output))
}

func TestFormatJSONEscapesStringsWithoutHTMLEscapes(t *testing.T) {
	t.Parallel()

	input := []byte("{\"s\":\"quote\\\" slash\\\\ backspace\\b formfeed\\f newline\\n carriage\\r tab\\t <>& snowman ☃ line\\u2028para\\u2029\"}")
	output, err := formatJSON(input, DefaultConfig())
	if err != nil {
		t.Fatalf("formatJSON returned error: %v", err)
	}

	expected := "{\n  \"s\": \"quote\\\" slash\\\\ backspace\\b formfeed\\f newline\\n carriage\\r tab\\t <>& snowman ☃ line\\u2028para\\u2029\"\n}\n"
	assertStringEqual(t, expected, string(output))
	assertJSONSemanticallyEqual(t, input, output)
}

func TestFormatJSONPreservesObjectOrderByDefault(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.ObjectExpand = ExpandNever

	output, err := formatJSON([]byte(`{"z":1,"a":2}`), cfg)
	if err != nil {
		t.Fatalf("formatJSON returned error: %v", err)
	}

	assertStringEqual(t, "{\"z\": 1, \"a\": 2}\n", string(output))
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

	assertStringEqual(t, "{ \"emptyObject\": { }, \"emptyArray\": [ ] }\n", string(output))
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
	assertStringEqual(t, expected, string(output))
}

func TestFormatJSONExpandNeverIgnoresPrintWidth(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.PrintWidth = 1
	cfg.ObjectExpand = ExpandNever
	cfg.ArrayExpand = ExpandNever

	output, err := formatJSON([]byte(`{"alpha":[1,2,3]}`), cfg)
	if err != nil {
		t.Fatalf("formatJSON returned error: %v", err)
	}

	assertStringEqual(t, "{\"alpha\": [1, 2, 3]}\n", string(output))
}
