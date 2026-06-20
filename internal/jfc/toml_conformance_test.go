package jfc

import (
	"strings"
	"testing"
)

func TestFormatTOMLAcceptsConformanceCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		{name: "scalars", input: "title=\"jfc\"\nactive=true\ncount=1\npi=3.14\n"},
		{name: "dates_and_times", input: "date=2026-06-20\noffset=2026-06-20T12:34:56-04:00\nlocal=12:34:56\n"},
		{name: "arrays", input: "items=[\"a\", \"b\", \"c\"]\npoints=[{x=1,y=2},{x=3,y=4}]\n"},
		{name: "inline_table", input: "metadata={name=\"jfc\", stable=true}\n"},
		{name: "dotted_keys", input: "server.alpha.ip=\"10.0.0.1\"\nserver.alpha.role=\"api\"\n"},
		{name: "tables", input: "[project]\nname=\"jfc\"\n\n[tool.jfc]\nline-length=80\n"},
		{name: "array_of_tables", input: "[[products]]\nname=\"Hammer\"\n\n[[products]]\nname=\"Nail\"\n"},
		{name: "comments", input: "# heading\nname=\"jfc\" # inline\n"},
		{name: "literal_strings", input: "path='C:\\Users\\node'\nregex='\\d+'\n"},
		{name: "multiline_strings", input: "basic=\"\"\"alpha=beta  \n# text\n\"\"\"\nliteral='''gamma=delta\t\n'''\n"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output, err := formatTOML([]byte(tc.input), DefaultConfig())
			if err != nil {
				t.Fatalf("formatTOML returned error: %v", err)
			}

			idempotent, err := formatTOML(output, DefaultConfig())
			if err != nil {
				t.Fatalf("formatTOML rejected its own output: %v\noutput:\n%s", err, output)
			}
			assertStringEqual(t, string(output), string(idempotent))
			assertTOMLSemanticallyEqual(t, []byte(tc.input), output)
		})
	}
}

func TestFormatTOMLRejectsConformanceCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		input   string
		message string
	}{
		{name: "missing_equals", input: "name \"jfc\"\n", message: "expected"},
		{name: "duplicate_key", input: "name=\"a\"\nname=\"b\"\n", message: "already defined"},
		{name: "invalid_boolean", input: "active=yes\n", message: "number"},
		{name: "invalid_date", input: "date=2026-99-99\n", message: "date"},
		{name: "unterminated_string", input: "name=\"jfc\n", message: "strings cannot have new lines"},
		{name: "unterminated_array", input: "items=[1,2\n", message: "expected"},
		{name: "unterminated_inline_table", input: "point={x=1\n", message: "expected"},
		{name: "invalid_table", input: "[tool\nname=\"jfc\"\n", message: "expected"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := formatTOML([]byte(tc.input), DefaultConfig())
			if err == nil {
				t.Fatal("expected formatTOML to reject input")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.message)) {
				t.Fatalf("expected error containing %q, got %v", tc.message, err)
			}
		})
	}
}
