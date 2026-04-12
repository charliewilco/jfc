package jfc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatJSONUsesTabsWhenExpanded(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.UseTabs = true
	cfg.TabWidth = 4
	cfg.ArrayExpand = ExpandAlways
	cfg.ObjectExpand = ExpandAlways

	output, err := formatJSON([]byte(`{"alpha":[1,{"beta":2}]}`), cfg)
	if err != nil {
		t.Fatalf("formatJSON returned error: %v", err)
	}

	expected := "{\n\t\"alpha\": [\n\t\t1,\n\t\t{\n\t\t\t\"beta\": 2\n\t\t}\n\t]\n}\n"
	if string(output) != expected {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestFormatJSONAppliesInlineSpacingOptions(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.ObjectExpand = ExpandNever
	cfg.SpaceWithinBraces = true
	cfg.SpaceWithinBrackets = true

	output, err := formatJSON([]byte(`{"alpha":[1,2]}`), cfg)
	if err != nil {
		t.Fatalf("formatJSON returned error: %v", err)
	}

	expected := "{ \"alpha\": [ 1, 2 ] }\n"
	if string(output) != expected {
		t.Fatalf("unexpected output:\n%s", output)
	}
}

func TestFormatJSONSortsKeysWhenEnabled(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.SortKeys = true
	cfg.ObjectExpand = ExpandNever

	output, err := formatJSON([]byte(`{"z":1,"a":2}`), cfg)
	if err != nil {
		t.Fatalf("formatJSON returned error: %v", err)
	}

	if got := string(output); got != "{\"a\": 2, \"z\": 1}\n" {
		t.Fatalf("unexpected output %q", got)
	}
}

func TestRunWriteDiscoversNearestConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	subdir := filepath.Join(root, "nested")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	config := "use_tabs = true\ntrailing_newline = true\nobject_expand = \"always\"\n"
	if err := os.WriteFile(filepath.Join(root, defaultConfigName), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	target := filepath.Join(subdir, "example.json")
	if err := os.WriteFile(target, []byte(`{"x":1}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--write", target}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitSuccess {
		t.Fatalf("Run exit code = %d, stderr = %s", exitCode, stderr.String())
	}

	formatted, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read formatted file: %v", err)
	}

	expected := "{\n\t\"x\": 1\n}\n"
	if string(formatted) != expected {
		t.Fatalf("unexpected file contents:\n%s", formatted)
	}
	if strings.TrimSpace(stdout.String()) != target {
		t.Fatalf("expected written path in stdout, got %q", stdout.String())
	}
}

func TestRunCheckReturnsNonZeroForUnformattedFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "example.json")
	if err := os.WriteFile(target, []byte(`{"x":1}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--check", target}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitDiff {
		t.Fatalf("Run exit code = %d, want %d, stderr = %s", exitCode, exitDiff, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != target {
		t.Fatalf("expected differing path in stdout, got %q", stdout.String())
	}
}

func TestRunReportsInvalidJSONWithLineAndColumn(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(nil, strings.NewReader("{\n\t\"x\":\n}"), &stdout, &stderr, func() (string, error) {
		return t.TempDir(), nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
	}
	if !strings.Contains(stderr.String(), "line 3, column 1") {
		t.Fatalf("expected line/column in stderr, got %q", stderr.String())
	}
}
