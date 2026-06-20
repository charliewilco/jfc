package jfc

import (
	"strings"
	"testing"
)

func TestFormatJSONAcceptsConformanceCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		{name: "empty_array", input: `[]`},
		{name: "empty_object", input: `{}`},
		{name: "top_level_string", input: `"ok"`},
		{name: "top_level_number", input: `-9223372036854775808`},
		{name: "top_level_false", input: `false`},
		{name: "nested_empty_values", input: `{"a":[],"b":{},"c":[{},[]]}`},
		{name: "all_scalar_values", input: `[null,true,false,0,-0,1.25,-1.25e+30]`},
		{name: "leading_trailing_whitespace", input: " \n\t {\"ok\":true} \r\n"},
		{name: "escaped_characters", input: `{"s":"quote\" slash\\ backspace\b formfeed\f newline\n carriage\r tab\t"}`},
		{name: "unicode_escape", input: `{"snowman":"\u2603","surrogate":"\uD834\uDD1E"}`},
		{name: "duplicate_keys", input: `{"a":1,"a":2}`},
		{name: "deep_nesting", input: `{"a":[{"b":[{"c":[1,2,3]}]}]}`},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output, err := formatJSON([]byte(tc.input), DefaultConfig())
			if err != nil {
				t.Fatalf("formatJSON returned error: %v", err)
			}

			idempotent, err := formatJSON(output, DefaultConfig())
			if err != nil {
				t.Fatalf("formatJSON rejected its own output: %v\noutput:\n%s", err, output)
			}
			assertStringEqual(t, string(output), string(idempotent))
			assertJSONSemanticallyEqual(t, []byte(tc.input), output)
		})
	}
}

func TestFormatJSONRejectsConformanceCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   string
		message string
	}{
		{name: "empty_input", input: ``, message: "unexpected end of input"},
		{name: "trailing_comma_array", input: `[1,]`, message: "unexpected character"},
		{name: "trailing_comma_object", input: `{"a":1,}`, message: "expected object key string"},
		{name: "missing_comma_array", input: `[1 2]`, message: "expected ',' or ']' in array"},
		{name: "missing_comma_object", input: `{"a":1 "b":2}`, message: "expected ',' or '}' in object"},
		{name: "single_quoted_string", input: `{'a':1}`, message: "expected object key string"},
		{name: "unquoted_key", input: `{a:1}`, message: "expected object key string"},
		{name: "leading_zero", input: `{"n": 01}`, message: "leading zeroes are not allowed"},
		{name: "plus_number", input: `+1`, message: "unexpected character"},
		{name: "missing_fraction_digit", input: `1.`, message: "fractional part requires at least one digit"},
		{name: "missing_exponent_digit", input: `1e`, message: "exponent requires at least one digit"},
		{name: "bad_literal", input: `tru`, message: "unexpected end of input"},
		{name: "unknown_literal", input: `undefined`, message: "unexpected character"},
		{name: "unterminated_string", input: `"abc`, message: "unterminated string"},
		{name: "raw_newline_in_string", input: "\"a\nb\"", message: "unterminated string"},
		{name: "raw_tab_in_string", input: "\"a\tb\"", message: "invalid string escape sequence"},
		{name: "invalid_escape", input: `"\x"`, message: "invalid string escape sequence"},
		{name: "invalid_unicode_escape", input: `"\u12xz"`, message: "invalid unicode escape"},
		{name: "trailing_content", input: `{} true`, message: "unexpected trailing content"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := formatJSON([]byte(tc.input), DefaultConfig())
			if err == nil {
				t.Fatal("expected formatJSON to reject input")
			}
			if !strings.Contains(err.Error(), tc.message) {
				t.Fatalf("expected error containing %q, got %v", tc.message, err)
			}
		})
	}
}
