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
	if stdout.String() != "{\"a\": 2, \"z\": 1}\n" {
		t.Fatalf("unexpected stdout %q", stdout.String())
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
	if string(contents) != `{"x":1}` {
		t.Fatalf("expected symlink target to remain unchanged, got %q", contents)
	}
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
	if string(contents) != "{\"x\": 1}\n" {
		t.Fatalf("unexpected target contents %q", contents)
	}
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
	if stdout.String() != "# Title\n" {
		t.Fatalf("unexpected stdout %q", stdout.String())
	}
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
	if stdout.String() != "{\"a\": 2, \"z\": 1}\n" {
		t.Fatalf("unexpected stdout %q", stdout.String())
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
