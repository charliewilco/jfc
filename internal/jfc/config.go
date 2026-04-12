package jfc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const defaultConfigName = "jfc.toml"

type ExpandMode string

const (
	ExpandAuto   ExpandMode = "auto"
	ExpandAlways ExpandMode = "always"
	ExpandNever  ExpandMode = "never"
)

type EndOfLine string

const (
	EndOfLineLF   EndOfLine = "lf"
	EndOfLineCRLF EndOfLine = "crlf"
	EndOfLineCR   EndOfLine = "cr"
)

type Config struct {
	UseTabs             bool
	TabWidth            int
	PrintWidth          int
	TrailingNewline     bool
	SortKeys            bool
	ArrayExpand         ExpandMode
	ObjectExpand        ExpandMode
	SpaceAfterColon     bool
	SpaceWithinBraces   bool
	SpaceWithinBrackets bool
	EndOfLine           EndOfLine
}

func DefaultConfig() Config {
	return Config{
		UseTabs:             false,
		TabWidth:            2,
		PrintWidth:          80,
		TrailingNewline:     true,
		SortKeys:            false,
		ArrayExpand:         ExpandAuto,
		ObjectExpand:        ExpandAuto,
		SpaceAfterColon:     true,
		SpaceWithinBraces:   false,
		SpaceWithinBrackets: false,
		EndOfLine:           EndOfLineLF,
	}
}

type configLoader struct {
	explicitPath string
	explicitCfg  *Config
	searchCache  map[string]*Config
}

func newConfigLoader(explicitPath string) *configLoader {
	return &configLoader{
		explicitPath: explicitPath,
		searchCache:  make(map[string]*Config),
	}
}

func (l *configLoader) forFile(path string) (Config, error) {
	if l.explicitPath != "" {
		return l.loadExplicit()
	}

	return l.loadDiscovered(filepath.Dir(path))
}

func (l *configLoader) forStdin(cwd string, stdinFilepath string) (Config, error) {
	if l.explicitPath != "" {
		return l.loadExplicit()
	}

	if stdinFilepath != "" {
		return l.loadDiscovered(filepath.Dir(stdinFilepath))
	}

	return l.loadDiscovered(cwd)
}

func (l *configLoader) loadExplicit() (Config, error) {
	if l.explicitCfg != nil {
		return *l.explicitCfg, nil
	}

	cfg, err := loadConfigFile(l.explicitPath)
	if err != nil {
		return Config{}, err
	}

	l.explicitCfg = &cfg
	return cfg, nil
}

func (l *configLoader) loadDiscovered(startDir string) (Config, error) {
	startDir = filepath.Clean(startDir)
	if cfg, ok := l.searchCache[startDir]; ok {
		if cfg == nil {
			return DefaultConfig(), nil
		}
		return *cfg, nil
	}

	configPath, found, err := findConfigPath(startDir)
	if err != nil {
		return Config{}, err
	}
	if !found {
		l.searchCache[startDir] = nil
		return DefaultConfig(), nil
	}

	cfg, err := loadConfigFile(configPath)
	if err != nil {
		return Config{}, err
	}

	l.searchCache[startDir] = &cfg
	return cfg, nil
}

func findConfigPath(startDir string) (string, bool, error) {
	dir := filepath.Clean(startDir)

	for {
		candidate := filepath.Join(dir, defaultConfigName)
		info, err := os.Stat(candidate)
		switch {
		case err == nil && !info.IsDir():
			return candidate, true, nil
		case err == nil && info.IsDir():
			return "", false, fmt.Errorf("%s is a directory, expected a TOML file", candidate)
		case errors.Is(err, os.ErrNotExist):
		case err != nil:
			return "", false, err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false, nil
		}
		dir = parent
	}
}

func loadConfigFile(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("load config %s: %w", path, err)
	}
	defer file.Close()

	cfg := DefaultConfig()
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	seen := make(map[string]struct{})

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(stripTOMLComment(scanner.Text()))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			return Config{}, fmt.Errorf("%s:%d: tables are not supported; use top-level key/value pairs only", path, lineNumber)
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Config{}, fmt.Errorf("%s:%d: expected key = value", path, lineNumber)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			return Config{}, fmt.Errorf("%s:%d: expected key = value", path, lineNumber)
		}
		if _, exists := seen[key]; exists {
			return Config{}, fmt.Errorf("%s:%d: duplicate key %q", path, lineNumber, key)
		}
		seen[key] = struct{}{}

		if err := applyConfigValue(&cfg, key, value); err != nil {
			return Config{}, fmt.Errorf("%s:%d: %w", path, lineNumber, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	return cfg, validateConfig(cfg)
}

func applyConfigValue(cfg *Config, key string, rawValue string) error {
	switch key {
	case "use_tabs":
		value, err := parseTOMLBool(rawValue)
		if err != nil {
			return err
		}
		cfg.UseTabs = value
	case "tab_width":
		value, err := parseTOMLInt(rawValue)
		if err != nil {
			return err
		}
		cfg.TabWidth = value
	case "print_width":
		value, err := parseTOMLInt(rawValue)
		if err != nil {
			return err
		}
		cfg.PrintWidth = value
	case "trailing_newline":
		value, err := parseTOMLBool(rawValue)
		if err != nil {
			return err
		}
		cfg.TrailingNewline = value
	case "sort_keys":
		value, err := parseTOMLBool(rawValue)
		if err != nil {
			return err
		}
		cfg.SortKeys = value
	case "array_expand":
		value, err := parseTOMLString(rawValue)
		if err != nil {
			return err
		}
		cfg.ArrayExpand = ExpandMode(strings.ToLower(value))
	case "object_expand":
		value, err := parseTOMLString(rawValue)
		if err != nil {
			return err
		}
		cfg.ObjectExpand = ExpandMode(strings.ToLower(value))
	case "space_after_colon":
		value, err := parseTOMLBool(rawValue)
		if err != nil {
			return err
		}
		cfg.SpaceAfterColon = value
	case "space_within_braces":
		value, err := parseTOMLBool(rawValue)
		if err != nil {
			return err
		}
		cfg.SpaceWithinBraces = value
	case "space_within_brackets":
		value, err := parseTOMLBool(rawValue)
		if err != nil {
			return err
		}
		cfg.SpaceWithinBrackets = value
	case "end_of_line":
		value, err := parseTOMLString(rawValue)
		if err != nil {
			return err
		}
		cfg.EndOfLine = EndOfLine(strings.ToLower(value))
	default:
		return fmt.Errorf("unknown config key %q", key)
	}

	return nil
}

func validateConfig(cfg Config) error {
	if cfg.TabWidth <= 0 {
		return fmt.Errorf("tab_width must be greater than zero")
	}
	if cfg.PrintWidth <= 0 {
		return fmt.Errorf("print_width must be greater than zero")
	}
	switch cfg.ArrayExpand {
	case ExpandAuto, ExpandAlways, ExpandNever:
	default:
		return fmt.Errorf("array_expand must be one of %q, %q, or %q", ExpandAuto, ExpandAlways, ExpandNever)
	}
	switch cfg.ObjectExpand {
	case ExpandAuto, ExpandAlways, ExpandNever:
	default:
		return fmt.Errorf("object_expand must be one of %q, %q, or %q", ExpandAuto, ExpandAlways, ExpandNever)
	}
	switch cfg.EndOfLine {
	case EndOfLineLF, EndOfLineCRLF, EndOfLineCR:
	default:
		return fmt.Errorf("end_of_line must be one of %q, %q, or %q", EndOfLineLF, EndOfLineCRLF, EndOfLineCR)
	}
	return nil
}

func parseTOMLBool(raw string) (bool, error) {
	switch raw {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("expected boolean, got %q", raw)
	}
}

func parseTOMLInt(raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("expected integer, got %q", raw)
	}
	return value, nil
}

func parseTOMLString(raw string) (string, error) {
	if len(raw) < 2 {
		return "", fmt.Errorf("expected quoted string, got %q", raw)
	}

	switch raw[0] {
	case '"':
		value, err := strconv.Unquote(raw)
		if err != nil {
			return "", fmt.Errorf("invalid quoted string %q", raw)
		}
		return value, nil
	case '\'':
		if raw[len(raw)-1] != '\'' {
			return "", fmt.Errorf("invalid literal string %q", raw)
		}
		return raw[1 : len(raw)-1], nil
	default:
		return "", fmt.Errorf("expected quoted string, got %q", raw)
	}
}

func stripTOMLComment(line string) string {
	var (
		inBasicString   bool
		inLiteralString bool
		escaped         bool
	)

	for i, r := range line {
		switch {
		case escaped:
			escaped = false
		case inBasicString && r == '\\':
			escaped = true
		case !inLiteralString && r == '"':
			inBasicString = !inBasicString
		case !inBasicString && r == '\'':
			inLiteralString = !inLiteralString
		case !inBasicString && !inLiteralString && r == '#':
			return line[:i]
		}
	}

	return line
}

func (c Config) indentUnit() string {
	if c.UseTabs {
		return "\t"
	}
	return strings.Repeat(" ", c.TabWidth)
}

func (c Config) endOfLineString() string {
	switch c.EndOfLine {
	case EndOfLineCRLF:
		return "\r\n"
	case EndOfLineCR:
		return "\r"
	default:
		return "\n"
	}
}
