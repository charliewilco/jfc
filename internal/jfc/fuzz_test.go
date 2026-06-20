package jfc

import "testing"

func FuzzFormatJSONPreservesSemanticsAndIsIdempotent(f *testing.F) {
	for _, seed := range []string{
		`null`,
		`true`,
		`123.45e-6`,
		`"hello\nworld"`,
		`[]`,
		`{}`,
		`{"z":1,"a":[3,2]}`,
		`{"unicode":"Jos\u00e9","escaped":"<>&","nested":{"items":[true,false,null]}}`,
		`[{"id":1},{"id":2,"tags":["a","b"]}]`,
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		cfg := DefaultConfig()
		cfg.SortKeys = true

		output, err := formatJSON([]byte(input), cfg)
		if err != nil {
			return
		}

		idempotent, err := formatJSON(output, cfg)
		if err != nil {
			t.Fatalf("formatJSON rejected its own output: %v\noutput:\n%s", err, output)
		}
		assertStringEqual(t, string(output), string(idempotent))
		assertJSONSemanticallyEqual(t, []byte(input), output)
	})
}

func FuzzFormatTOMLPreservesSemanticsAndIsIdempotent(f *testing.F) {
	for _, seed := range []string{
		`title="jfc"`,
		"answer=42\npi=3.14159\nactive=true\n",
		"date=2026-06-20\ntime=07:32:00Z\n",
		"items=[\"a\", \"b\", {nested=true}]\n",
		"point={x=1, y=2}\n",
		"servers.alpha.ip=\"10.0.0.1\"\n",
		"[tool]\nname=\"jfc\"\nitems=[\"json\", \"toml\"]\n",
		"basic=\"\"\"alpha=beta  \n# not a comment\n\"\"\"\n",
		"literal='''gamma=delta\t\n# also not a comment\n'''\n",
		"commented=\"value\" # comment has a=b\n",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		output, err := formatTOML([]byte(input), DefaultConfig())
		if err != nil {
			return
		}

		idempotent, err := formatTOML(output, DefaultConfig())
		if err != nil {
			t.Fatalf("formatTOML rejected its own output: %v\noutput:\n%s", err, output)
		}
		assertStringEqual(t, string(output), string(idempotent))
		assertTOMLSemanticallyEqual(t, []byte(input), output)
	})
}
