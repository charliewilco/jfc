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
