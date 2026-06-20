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
	assertStringEqual(t, expected, string(output))
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
	assertStringEqual(t, expected, string(output))
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

	assertStringEqual(t, "{\"a\": 2, \"z\": 1}\n", string(output))
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
	assertStringEqual(t, expected, string(formatted))
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

func TestRunDiffPrintsUnifiedDiffForUnformattedFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "example.json")
	if err := os.WriteFile(target, []byte(`{"x":1}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--diff", target}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitDiff {
		t.Fatalf("Run exit code = %d, want %d, stderr = %s", exitCode, exitDiff, stderr.String())
	}

	expected := strings.Join([]string{
		"--- " + target,
		"+++ " + target,
		"@@ -1,1 +1,1 @@",
		`-{"x":1}`,
		`\ No newline at end of file`,
		`+{"x": 1}`,
		"",
	}, "\n")
	assertStringEqual(t, expected, stdout.String())
}

func TestRunCheckDiffPrintsUnifiedDiffForUnformattedFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "example.json")
	if err := os.WriteFile(target, []byte(`{"x":1}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--check", "--diff", target}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitDiff {
		t.Fatalf("Run exit code = %d, want %d, stderr = %s", exitCode, exitDiff, stderr.String())
	}
	if !strings.Contains(stdout.String(), "--- "+target+"\n+++ "+target+"\n") {
		t.Fatalf("expected unified diff in stdout, got %q", stdout.String())
	}
	if strings.Contains(stdout.String(), target+"\n--- ") {
		t.Fatalf("expected diff output without separate path listing, got %q", stdout.String())
	}
}

func TestRunDiffPrintsUnifiedDiffForStdinFilepath(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(
		[]string{"--diff", "--stdin-filepath", "README.md"},
		strings.NewReader("# Title\r\n   \r\n"),
		&stdout,
		&stderr,
		func() (string, error) {
			return t.TempDir(), nil
		},
	)
	if exitCode != exitDiff {
		t.Fatalf("Run exit code = %d, want %d, stderr = %s", exitCode, exitDiff, stderr.String())
	}
	if !strings.Contains(stdout.String(), "--- README.md\n+++ README.md\n") {
		t.Fatalf("expected stdin filepath labels in diff, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "-   ") || !strings.Contains(stdout.String(), "+# Title") {
		t.Fatalf("expected markdown changes in diff, got %q", stdout.String())
	}
}

func TestRunCheckDiffPrintsUnifiedDiffForStdinFilepath(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(
		[]string{"--check", "--diff", "--stdin-filepath", "README.md"},
		strings.NewReader("# Title\r\n   \r\n"),
		&stdout,
		&stderr,
		func() (string, error) {
			return t.TempDir(), nil
		},
	)
	if exitCode != exitDiff {
		t.Fatalf("Run exit code = %d, want %d, stderr = %s", exitCode, exitDiff, stderr.String())
	}
	if !strings.Contains(stdout.String(), "--- README.md\n+++ README.md\n") {
		t.Fatalf("expected stdin filepath labels in diff, got %q", stdout.String())
	}
}

func TestRunCheckReturnsErrorWhenAnyTargetCannotFormat(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	invalid := filepath.Join(root, "invalid.json")
	changed := filepath.Join(root, "changed.json")
	if err := os.WriteFile(invalid, []byte(`{"bad":}`), 0o644); err != nil {
		t.Fatalf("write invalid json: %v", err)
	}
	if err := os.WriteFile(changed, []byte(`{"x":1}`), 0o644); err != nil {
		t.Fatalf("write changed json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--check", root}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d, stdout = %q, stderr = %q", exitCode, exitError, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), changed) {
		t.Fatalf("expected changed file in stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), invalid) {
		t.Fatalf("expected invalid file error in stderr, got %q", stderr.String())
	}
}

func TestRunCheckSkipsIgnoredConfigPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	ignoredDir := filepath.Join(root, "dist", "ignored.json")
	ignoredGenerated := filepath.Join(root, "api.generated.json")
	checked := filepath.Join(root, "checked.json")
	if err := os.MkdirAll(filepath.Dir(ignoredDir), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	config := "ignore = [\"dist\", \"*.generated.json\"]\n"
	if err := os.WriteFile(filepath.Join(root, defaultConfigName), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	for _, path := range []string{ignoredDir, ignoredGenerated, checked} {
		if err := os.WriteFile(path, []byte(`{"x":1}`), 0o644); err != nil {
			t.Fatalf("write json: %v", err)
		}
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--check", root}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitDiff {
		t.Fatalf("Run exit code = %d, want %d, stdout = %q, stderr = %q", exitCode, exitDiff, stdout.String(), stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, checked) {
		t.Fatalf("expected checked path in stdout, got %q", output)
	}
	if strings.Contains(output, ignoredDir) || strings.Contains(output, ignoredGenerated) {
		t.Fatalf("expected ignored paths to be skipped, got %q", output)
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunCheckSkipsStandardIgnoreSources(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	paths := map[string]string{
		".ignore":                    "dist/\n*.generated.json\n!keep.generated.json\n",
		".gitignore":                 "vendor/\n",
		".git/info/exclude":          "excluded.json\n",
		".jfcignore":                 "not-a-jfc-ignore-file.json\n",
		"dist/ignored.json":          `{"ignored":"dist"}`,
		"vendor/ignored.json":        `{"ignored":"vendor"}`,
		"api.generated.json":         `{"ignored":"generated"}`,
		"excluded.json":              `{"ignored":"exclude"}`,
		"keep.generated.json":        `{"kept":"negated"}`,
		"checked.json":               `{"checked":true}`,
		"not-a-jfc-ignore-file.json": `{"checked":"jfcignore is not supported"}`,
	}
	for name, contents := range paths {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--check", root}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitDiff {
		t.Fatalf("Run exit code = %d, want %d, stdout = %q, stderr = %q", exitCode, exitDiff, stdout.String(), stderr.String())
	}

	output := stdout.String()
	for _, name := range []string{"checked.json", "keep.generated.json", "not-a-jfc-ignore-file.json"} {
		if !strings.Contains(output, filepath.Join(root, name)) {
			t.Fatalf("expected %s in stdout, got %q", name, output)
		}
	}
	for _, name := range []string{"dist/ignored.json", "vendor/ignored.json", "api.generated.json", "excluded.json"} {
		if strings.Contains(output, filepath.Join(root, name)) {
			t.Fatalf("expected %s to be ignored, got %q", name, output)
		}
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunListDifferentTraversesDirectories(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	formatted := filepath.Join(root, "formatted.json")
	unformatted := filepath.Join(root, "nested", "example.json")
	if err := os.MkdirAll(filepath.Dir(unformatted), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(formatted, []byte("{\"ok\": true}\n"), 0o644); err != nil {
		t.Fatalf("write formatted file: %v", err)
	}
	if err := os.WriteFile(unformatted, []byte(`{"x":1}`), 0o644); err != nil {
		t.Fatalf("write unformatted file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--list-different", root}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitDiff {
		t.Fatalf("Run exit code = %d, want %d, stderr = %s", exitCode, exitDiff, stderr.String())
	}

	lines := strings.Fields(stdout.String())
	if len(lines) != 1 || lines[0] != unformatted {
		t.Fatalf("expected only differing file path, got %q", stdout.String())
	}
}

func TestRunCheckAcceptsGlobInput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	first := filepath.Join(root, "a.json")
	second := filepath.Join(root, "nested", "b.json")
	if err := os.MkdirAll(filepath.Dir(second), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(first, []byte("{\"a\": 1}\n"), 0o644); err != nil {
		t.Fatalf("write formatted file: %v", err)
	}
	if err := os.WriteFile(second, []byte(`{"b":1}`), 0o644); err != nil {
		t.Fatalf("write unformatted file: %v", err)
	}

	pattern := filepath.Join(root, "*", "*.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--check", pattern}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitDiff {
		t.Fatalf("Run exit code = %d, want %d, stderr = %s", exitCode, exitDiff, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != second {
		t.Fatalf("expected glob result %q, got %q", second, stdout.String())
	}
}

func TestRunCheckAcceptsRecursiveGlobInput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	direct := filepath.Join(root, "fixtures", "direct.jsonc")
	nested := filepath.Join(root, "fixtures", "a", "b", "nested.jsonc")
	if err := os.MkdirAll(filepath.Dir(nested), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(direct, []byte("{\"direct\": true}\n"), 0o644); err != nil {
		t.Fatalf("write direct file: %v", err)
	}
	if err := os.WriteFile(nested, []byte(`{"nested":true}`), 0o644); err != nil {
		t.Fatalf("write nested file: %v", err)
	}

	pattern := filepath.Join(root, "fixtures", "**", "*.jsonc")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--check", pattern}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitDiff {
		t.Fatalf("Run exit code = %d, want %d, stderr = %s", exitCode, exitDiff, stderr.String())
	}

	lines := strings.Fields(stdout.String())
	if len(lines) != 1 || lines[0] != nested {
		t.Fatalf("expected only nested file to differ, got %q", stdout.String())
	}
}

func TestRunCheckRecursiveGlobSkipsGitDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gitObject := filepath.Join(root, ".git", "objects", "internal.json")
	workingFile := filepath.Join(root, "fixtures", "working.json")
	if err := os.MkdirAll(filepath.Dir(gitObject), 0o755); err != nil {
		t.Fatalf("mkdir .git object dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(workingFile), 0o755); err != nil {
		t.Fatalf("mkdir working dir: %v", err)
	}
	if err := os.WriteFile(gitObject, []byte(`{"git":true}`), 0o644); err != nil {
		t.Fatalf("write git object: %v", err)
	}
	if err := os.WriteFile(workingFile, []byte(`{"working":true}`), 0o644); err != nil {
		t.Fatalf("write working file: %v", err)
	}

	pattern := filepath.Join(root, "**", "*.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--check", pattern}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitDiff {
		t.Fatalf("Run exit code = %d, want %d, stderr = %s", exitCode, exitDiff, stderr.String())
	}

	output := stdout.String()
	if strings.Contains(output, ".git") {
		t.Fatalf("recursive glob should skip .git, got %q", output)
	}
	if !strings.Contains(output, workingFile) {
		t.Fatalf("expected working file in output, got %q", output)
	}
}

func TestRunCheckRejectsUnmatchedRecursiveGlob(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	pattern := filepath.Join(root, "fixtures", "**", "*.jsonc")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--check", pattern}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
	}
	if !strings.Contains(stderr.String(), "did not match any files") {
		t.Fatalf("expected unmatched glob error, got %q", stderr.String())
	}
}

func TestRecursiveGlobRoot(t *testing.T) {
	t.Parallel()

	if got := recursiveGlobRoot(filepath.Join("fixtures", "**", "*.jsonc")); got != "fixtures" {
		t.Fatalf("recursiveGlobRoot relative = %q, want fixtures", got)
	}

	absolutePattern := filepath.Join(string(filepath.Separator), "tmp", "**", "*.jsonc")
	if got := recursiveGlobRoot(absolutePattern); got != filepath.Join(string(filepath.Separator), "tmp") {
		t.Fatalf("recursiveGlobRoot absolute = %q, want /tmp", got)
	}
}

func TestCollectTargetsIncludesAllSupportedExtensions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	names := []string{
		"data.json",
		"settings.jsonc",
		"events.jsonl",
		"events.ndjson",
		"config.yaml",
		"config.yml",
		"tool.toml",
		"README.md",
		"README.markdown",
		"main.go",
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(root, name), []byte("{}\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	targets, err := collectTargets([]string{root})
	if err != nil {
		t.Fatalf("collectTargets returned error: %v", err)
	}
	if len(targets) != len(names)-1 {
		t.Fatalf("expected %d supported files, got %d: %v", len(names)-1, len(targets), targets)
	}
	for _, target := range targets {
		if filepath.Base(target) == "main.go" {
			t.Fatalf("unsupported file included in targets: %v", targets)
		}
	}
}

func TestRunUsesStdinFilepathForConfigDiscovery(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	subdir := filepath.Join(root, "nested")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	config := "sort_keys = true\nobject_expand = \"never\"\n"
	if err := os.WriteFile(filepath.Join(root, defaultConfigName), []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	target := filepath.Join(subdir, "stdin.json")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--stdin-filepath", target}, strings.NewReader(`{"z":1,"a":2}`), &stdout, &stderr, func() (string, error) {
		return subdir, nil
	})
	if exitCode != exitSuccess {
		t.Fatalf("Run exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	assertStringEqual(t, "{\"a\": 2, \"z\": 1}\n", stdout.String())
}

func TestRunInitCreatesMinimalConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"init"}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitSuccess {
		t.Fatalf("Run exit code = %d, stderr = %s", exitCode, stderr.String())
	}

	configPath := filepath.Join(root, defaultConfigName)
	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	assertStringEqual(t, initConfigContents, string(contents))
	if strings.TrimSpace(stdout.String()) != configPath {
		t.Fatalf("expected created path in stdout, got %q", stdout.String())
	}
	if stderr.String() != "" {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunInitRefusesToOverwriteConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	configPath := filepath.Join(root, defaultConfigName)
	original := []byte("print_width = 100\n")
	if err := os.WriteFile(configPath, original, 0o644); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"init"}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
	}
	if stdout.String() != "" {
		t.Fatalf("expected empty stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "already exists") {
		t.Fatalf("expected existing config error, got %q", stderr.String())
	}

	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !bytes.Equal(contents, original) {
		t.Fatalf("existing config was overwritten: %q", contents)
	}
}

func TestRunInitRejectsArguments(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"init", "config/jfc.toml"}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return t.TempDir(), nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
	}
	if !strings.Contains(stderr.String(), "init does not accept arguments") {
		t.Fatalf("expected init argument error, got %q", stderr.String())
	}
}

func TestRunWriteTraversesSupportedFormats(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	files := map[string]string{
		"data.json":       `{"x":1}`,
		"settings.jsonc":  "{\n// keep\n\"x\": 1,\n}\n",
		"events.jsonl":    "{\"x\":1}\n",
		"config.yaml":     "root:\n  child: value\n",
		"tool.toml":       "name=\"jfc\"\n",
		"README.markdown": "# Title\r\n   \r\n",
	}
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(contents), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write unsupported file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--write", root}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitSuccess {
		t.Fatalf("Run exit code = %d, stderr = %s", exitCode, stderr.String())
	}

	for _, name := range []string{"data.json", "settings.jsonc", "events.jsonl", "tool.toml", "README.markdown"} {
		if !strings.Contains(stdout.String(), filepath.Join(root, name)) {
			t.Fatalf("expected %s in stdout, got %q", name, stdout.String())
		}
	}
	if strings.Contains(stdout.String(), "main.go") {
		t.Fatalf("unsupported file should not be traversed, got %q", stdout.String())
	}
}

func TestRunWriteSkipsSymlinksDuringDirectoryTraversal(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.json")
	if err := os.WriteFile(outside, []byte(`{"x":1}`), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	link := filepath.Join(root, "linked.json")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--write", root}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
	}
	if !strings.Contains(stderr.String(), "no supported files found") {
		t.Fatalf("expected no supported files error, got %q", stderr.String())
	}

	contents, err := os.ReadFile(outside)
	if err != nil {
		t.Fatalf("read outside file: %v", err)
	}
	assertStringEqual(t, `{"x":1}`, string(contents))
}

func TestRunWriteExplicitSymlinkUpdatesTargetAndPreservesLink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "target.json")
	if err := os.WriteFile(target, []byte(`{"x":1}`), 0o640); err != nil {
		t.Fatalf("write target file: %v", err)
	}
	link := filepath.Join(root, "linked.json")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--write", link}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitSuccess {
		t.Fatalf("Run exit code = %d, stderr = %s", exitCode, stderr.String())
	}

	linkInfo, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat link: %v", err)
	}
	if linkInfo.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected %s to remain a symlink", link)
	}
	contents, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target file: %v", err)
	}
	assertStringEqual(t, "{\"x\": 1}\n", string(contents))
	targetInfo, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat target: %v", err)
	}
	if targetInfo.Mode().Perm() != 0o640 {
		t.Fatalf("expected target mode 0640, got %v", targetInfo.Mode().Perm())
	}
}

func TestRunWritePreservesFileMode(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "example.json")
	if err := os.WriteFile(target, []byte(`{"x":1}`), 0o600); err != nil {
		t.Fatalf("write target file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--write", target}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitSuccess {
		t.Fatalf("Run exit code = %d, stderr = %s", exitCode, stderr.String())
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat target: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected mode 0600, got %v", info.Mode().Perm())
	}
}

func TestWriteFileAtomicallyReportsMissingPath(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing", "example.json")
	err := writeFileAtomically(path, []byte("{}\n"))
	if err == nil {
		t.Fatal("expected missing path error")
	}
}

func TestRunStdinFilepathSelectsMarkdownFormatter(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run(
		[]string{"--stdin-filepath", "README.md"},
		strings.NewReader("# Title\r\n   \r\n"),
		&stdout,
		&stderr,
		func() (string, error) {
			return t.TempDir(), nil
		},
	)
	if exitCode != exitSuccess {
		t.Fatalf("Run exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	assertStringEqual(t, "# Title\n", stdout.String())
}

func TestRunRejectsUnsupportedExplicitFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "main.go")
	if err := os.WriteFile(target, []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write unsupported file: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--check", target}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
	}
	if !strings.Contains(stderr.String(), "not a supported file") {
		t.Fatalf("expected supported file error, got %q", stderr.String())
	}
}

func TestRunExplicitConfigOverridesDiscovery(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	subdir := filepath.Join(root, "nested")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	localConfig := "sort_keys = false\nobject_expand = \"never\"\n"
	explicitConfig := filepath.Join(root, "custom.toml")
	if err := os.WriteFile(filepath.Join(root, defaultConfigName), []byte(localConfig), 0o644); err != nil {
		t.Fatalf("write discovered config: %v", err)
	}
	if err := os.WriteFile(explicitConfig, []byte("sort_keys = true\nobject_expand = \"never\"\n"), 0o644); err != nil {
		t.Fatalf("write explicit config: %v", err)
	}

	target := filepath.Join(subdir, "example.json")
	if err := os.WriteFile(target, []byte(`{"z":1,"a":2}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--config", explicitConfig, target}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitSuccess {
		t.Fatalf("Run exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	assertStringEqual(t, "{\"a\": 2, \"z\": 1}\n", stdout.String())
}

func TestRunReportsMissingExplicitConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	missingConfig := filepath.Join(root, "missing.toml")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--config", missingConfig}, strings.NewReader(`{"x":1}`), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
	}
	if stdout.String() != "" {
		t.Fatalf("expected no stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "load config") || !strings.Contains(stderr.String(), missingConfig) {
		t.Fatalf("expected missing explicit config error, got %q", stderr.String())
	}
}

func TestRunReportsInvalidExplicitConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	configPath := filepath.Join(root, "invalid.toml")
	target := filepath.Join(root, "example.json")
	if err := os.WriteFile(configPath, []byte("tab_width = 0\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(target, []byte(`{"x":1}`), 0o644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--config", configPath, "--check", target}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
	}
	if stdout.String() != "" {
		t.Fatalf("expected no stdout, got %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "tab_width must be greater than zero") {
		t.Fatalf("expected invalid explicit config error, got %q", stderr.String())
	}
}

func TestRunReportsBadDiscoveredConfigForFileAndStdin(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		run  func(t *testing.T, root string) (int, string, string)
	}{
		{
			name: "file",
			run: func(t *testing.T, root string) (int, string, string) {
				target := filepath.Join(root, "nested", "example.json")
				if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
				if err := os.WriteFile(target, []byte(`{"x":1}`), 0o644); err != nil {
					t.Fatalf("write json: %v", err)
				}

				var stdout bytes.Buffer
				var stderr bytes.Buffer
				exitCode := Run([]string{"--check", target}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
					return root, nil
				})
				return exitCode, stdout.String(), stderr.String()
			},
		},
		{
			name: "stdin",
			run: func(t *testing.T, root string) (int, string, string) {
				stdinPath := filepath.Join(root, "nested", "stdin.json")
				if err := os.MkdirAll(filepath.Dir(stdinPath), 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}

				var stdout bytes.Buffer
				var stderr bytes.Buffer
				exitCode := Run([]string{"--stdin-filepath", stdinPath}, strings.NewReader(`{"x":1}`), &stdout, &stderr, func() (string, error) {
					return root, nil
				})
				return exitCode, stdout.String(), stderr.String()
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, defaultConfigName), []byte("print_width = 0\n"), 0o644); err != nil {
				t.Fatalf("write config: %v", err)
			}

			exitCode, stdout, stderr := tc.run(t, root)
			if exitCode != exitError {
				t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
			}
			if stdout != "" {
				t.Fatalf("expected no stdout, got %q", stdout)
			}
			if !strings.Contains(stderr, "print_width must be greater than zero") {
				t.Fatalf("expected discovered config error, got %q", stderr)
			}
		})
	}
}

func TestRunHelpReturnsSuccess(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--help"}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return t.TempDir(), nil
	})
	if exitCode != exitSuccess {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitSuccess)
	}
	if !strings.Contains(stderr.String(), "Usage: jfc") {
		t.Fatalf("expected usage output, got %q", stderr.String())
	}
}

func TestRunRejectsMultipleFilesWithoutMode(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	first := filepath.Join(root, "a.json")
	second := filepath.Join(root, "b.json")
	if err := os.WriteFile(first, []byte("{\"a\": 1}\n"), 0o644); err != nil {
		t.Fatalf("write first json: %v", err)
	}
	if err := os.WriteFile(second, []byte("{\"b\": 1}\n"), 0o644); err != nil {
		t.Fatalf("write second json: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{first, second}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return root, nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
	}
	if !strings.Contains(stderr.String(), "multiple file arguments require") {
		t.Fatalf("expected multiple file error, got %q", stderr.String())
	}
}

func TestRunRejectsWriteWithStdin(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--write"}, strings.NewReader(`{"x":1}`), &stdout, &stderr, func() (string, error) {
		return t.TempDir(), nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
	}
	if !strings.Contains(stderr.String(), "--write cannot be used with stdin") {
		t.Fatalf("expected stdin write error, got %q", stderr.String())
	}
}

func TestRunRejectsMutuallyExclusiveModes(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"--write", "--check", "example.json"}, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
		return t.TempDir(), nil
	})
	if exitCode != exitError {
		t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
	}
	if !strings.Contains(stderr.String(), "mutually exclusive") {
		t.Fatalf("expected mutually exclusive flags error, got %q", stderr.String())
	}
}

func TestRunRejectsDiffWithWriteOrListDifferent(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{
		{"--write", "--diff", "example.json"},
		{"--list-different", "--diff", "example.json"},
	} {
		args := args
		t.Run(strings.Join(args[:2], "_"), func(t *testing.T) {
			t.Parallel()

			var stdout bytes.Buffer
			var stderr bytes.Buffer
			exitCode := Run(args, bytes.NewReader(nil), &stdout, &stderr, func() (string, error) {
				return t.TempDir(), nil
			})
			if exitCode != exitError {
				t.Fatalf("Run exit code = %d, want %d", exitCode, exitError)
			}
			if !strings.Contains(stderr.String(), "--diff can be used alone or with --check") {
				t.Fatalf("expected diff mode error, got %q", stderr.String())
			}
		})
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
