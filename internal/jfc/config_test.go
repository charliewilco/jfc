package jfc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigFileParsesSupportedSchema(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, defaultConfigName)
	contents := strings.Join([]string{
		"use_tabs = true",
		"tab_width = 4",
		"print_width = 100",
		"trailing_newline = false",
		"sort_keys = true",
		"array_expand = \"never\"",
		"object_expand = 'always'",
		"space_after_colon = false",
		"space_within_braces = true",
		"space_within_brackets = true",
		"end_of_line = \"crlf\" # normalize output for Windows",
	}, "\n")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfigFile(path)
	if err != nil {
		t.Fatalf("loadConfigFile returned error: %v", err)
	}

	if !cfg.UseTabs || cfg.TabWidth != 4 || cfg.PrintWidth != 100 || cfg.TrailingNewline || !cfg.SortKeys {
		t.Fatalf("unexpected scalar config values: %+v", cfg)
	}
	if cfg.ArrayExpand != ExpandNever || cfg.ObjectExpand != ExpandAlways {
		t.Fatalf("unexpected expand modes: %+v", cfg)
	}
	if cfg.SpaceAfterColon || !cfg.SpaceWithinBraces || !cfg.SpaceWithinBrackets {
		t.Fatalf("unexpected spacing config values: %+v", cfg)
	}
	if cfg.EndOfLine != EndOfLineCRLF {
		t.Fatalf("unexpected end_of_line: %q", cfg.EndOfLine)
	}
}

func TestLoadConfigFileUsesTOMLSyntax(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, defaultConfigName)
	contents := strings.Join([]string{
		"tab_width = 1_0",
		"print_width = +120",
		"array_expand = \"AUTO\"",
		"end_of_line = '''lf'''",
	}, "\n")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := loadConfigFile(path)
	if err != nil {
		t.Fatalf("loadConfigFile returned error: %v", err)
	}
	if cfg.TabWidth != 10 || cfg.PrintWidth != 120 {
		t.Fatalf("expected TOML integer syntax to decode, got %+v", cfg)
	}
	if cfg.ArrayExpand != ExpandAuto || cfg.EndOfLine != EndOfLineLF {
		t.Fatalf("expected string values to normalize, got %+v", cfg)
	}
}

func TestLoadConfigFileRejectsDuplicateKey(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, defaultConfigName)
	if err := os.WriteFile(path, []byte("tab_width = 2\ntab_width = 4\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := loadConfigFile(path)
	if err == nil || !strings.Contains(err.Error(), "already defined") {
		t.Fatalf("expected duplicate key error, got %v", err)
	}
}

func TestLoadConfigFileRejectsUnknownKey(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, defaultConfigName)
	if err := os.WriteFile(path, []byte("semi = false\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := loadConfigFile(path)
	if err == nil || !strings.Contains(err.Error(), "unknown config key") {
		t.Fatalf("expected unknown key error, got %v", err)
	}
}

func TestLoadConfigFileRejectsInvalidValue(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, defaultConfigName)
	if err := os.WriteFile(path, []byte("tab_width = 0\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := loadConfigFile(path)
	if err == nil || !strings.Contains(err.Error(), "tab_width must be greater than zero") {
		t.Fatalf("expected invalid value error, got %v", err)
	}
}

func TestLoadConfigFileAllowsHashInsideStrings(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, defaultConfigName)
	contents := "object_expand = \"always#still-a-string\"\n"
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := loadConfigFile(path)
	if err == nil || !strings.Contains(err.Error(), "object_expand must be one of") {
		t.Fatalf("expected full string to be parsed before validation, got %v", err)
	}
}

func TestLoadConfigFileRejectsTables(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, defaultConfigName)
	if err := os.WriteFile(path, []byte("[format]\nuse_tabs = true\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := loadConfigFile(path)
	if err == nil || !strings.Contains(err.Error(), "unknown config key") {
		t.Fatalf("expected table error, got %v", err)
	}
}

func TestFindConfigPathWalksUpward(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	subdir := filepath.Join(root, "one", "two")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	configPath := filepath.Join(root, defaultConfigName)
	if err := os.WriteFile(configPath, []byte("tab_width = 2\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	foundPath, found, err := findConfigPath(subdir)
	if err != nil {
		t.Fatalf("findConfigPath returned error: %v", err)
	}
	if !found || foundPath != configPath {
		t.Fatalf("expected %q, got found=%v path=%q", configPath, found, foundPath)
	}
}

func TestFindConfigPathReturnsNotFoundWhenMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	foundPath, found, err := findConfigPath(root)
	if err != nil {
		t.Fatalf("findConfigPath returned error: %v", err)
	}
	if found || foundPath != "" {
		t.Fatalf("expected no config, got found=%v path=%q", found, foundPath)
	}
}
