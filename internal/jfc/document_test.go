package jfc

import "testing"

func TestDetectFormatSupportsConfiguredExtensions(t *testing.T) {
	t.Parallel()

	cases := map[string]FormatKind{
		"data.json":       FormatJSON,
		"settings.jsonc":  FormatJSONC,
		"events.jsonl":    FormatJSONL,
		"events.ndjson":   FormatJSONL,
		"config.yaml":     FormatYAML,
		"config.yml":      FormatYAML,
		"jfc.toml":        FormatTOML,
		"README.md":       FormatMarkdown,
		"README.markdown": FormatMarkdown,
		"document.xml":    FormatXML,
		"vector.svg":      FormatXML,
		"Info.plist":      FormatXML,
		"View.xib":        FormatXML,
		"Main.storyboard": FormatXML,
		"App.csproj":      FormatXML,
		"App.vbproj":      FormatXML,
		"App.fsproj":      FormatXML,
		"Build.props":     FormatXML,
		"Build.targets":   FormatXML,
		"artifact.pom":    FormatXML,
		"data.csv":        FormatCSV,
		"data.tsv":        FormatTSV,
		".env":            FormatDotenv,
		".env.local":      FormatDotenv,
		"example.env":     FormatDotenv,
		"main.tf":         FormatHCL,
		"vars.tfvars":     FormatHCL,
		"policy.hcl":      FormatHCL,
		"job.nomad":       FormatHCL,
	}

	for path, want := range cases {
		got, ok := detectFormat(path)
		if !ok {
			t.Fatalf("detectFormat(%q) returned unsupported", path)
		}
		if got != want {
			t.Fatalf("detectFormat(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestDetectFormatRejectsUnsupportedExtensions(t *testing.T) {
	t.Parallel()

	if got, ok := detectFormat("main.go"); ok {
		t.Fatalf("detectFormat returned %q for unsupported file", got)
	}
}
