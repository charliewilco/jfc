package jfc

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
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

type configFile struct {
	UseTabs             *bool   `toml:"use_tabs"`
	TabWidth            *int    `toml:"tab_width"`
	PrintWidth          *int    `toml:"print_width"`
	TrailingNewline     *bool   `toml:"trailing_newline"`
	SortKeys            *bool   `toml:"sort_keys"`
	ArrayExpand         *string `toml:"array_expand"`
	ObjectExpand        *string `toml:"object_expand"`
	SpaceAfterColon     *bool   `toml:"space_after_colon"`
	SpaceWithinBraces   *bool   `toml:"space_within_braces"`
	SpaceWithinBrackets *bool   `toml:"space_within_brackets"`
	EndOfLine           *string `toml:"end_of_line"`
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
	input, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("load config %s: %w", path, err)
	}
	if err := rejectConfigTables(path, input); err != nil {
		return Config{}, err
	}

	cfg := DefaultConfig()
	var raw configFile
	decoder := toml.NewDecoder(bytes.NewReader(input))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return Config{}, formatConfigDecodeError(path, err)
	}

	applyConfigFile(&cfg, raw)
	return cfg, validateConfig(cfg)
}

func rejectConfigTables(path string, input []byte) error {
	var raw map[string]any
	if err := toml.Unmarshal(input, &raw); err != nil {
		return formatConfigDecodeError(path, err)
	}
	for key, value := range raw {
		if isTOMLTableValue(value) {
			return fmt.Errorf("%s: tables are not supported; use top-level key/value pairs only near %q", path, key)
		}
	}
	return nil
}

func isTOMLTableValue(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		return true
	case []map[string]any:
		return true
	case []any:
		for _, item := range typed {
			if isTOMLTableValue(item) {
				return true
			}
		}
	}
	return false
}

func applyConfigFile(cfg *Config, raw configFile) {
	if raw.UseTabs != nil {
		cfg.UseTabs = *raw.UseTabs
	}
	if raw.TabWidth != nil {
		cfg.TabWidth = *raw.TabWidth
	}
	if raw.PrintWidth != nil {
		cfg.PrintWidth = *raw.PrintWidth
	}
	if raw.TrailingNewline != nil {
		cfg.TrailingNewline = *raw.TrailingNewline
	}
	if raw.SortKeys != nil {
		cfg.SortKeys = *raw.SortKeys
	}
	if raw.ArrayExpand != nil {
		cfg.ArrayExpand = ExpandMode(strings.ToLower(*raw.ArrayExpand))
	}
	if raw.ObjectExpand != nil {
		cfg.ObjectExpand = ExpandMode(strings.ToLower(*raw.ObjectExpand))
	}
	if raw.SpaceAfterColon != nil {
		cfg.SpaceAfterColon = *raw.SpaceAfterColon
	}
	if raw.SpaceWithinBraces != nil {
		cfg.SpaceWithinBraces = *raw.SpaceWithinBraces
	}
	if raw.SpaceWithinBrackets != nil {
		cfg.SpaceWithinBrackets = *raw.SpaceWithinBrackets
	}
	if raw.EndOfLine != nil {
		cfg.EndOfLine = EndOfLine(strings.ToLower(*raw.EndOfLine))
	}
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

func formatConfigDecodeError(path string, err error) error {
	var strictErr *toml.StrictMissingError
	if errors.As(err, &strictErr) && len(strictErr.Errors) > 0 {
		key := strictErr.Errors[0].Key()
		if len(key) > 0 {
			return fmt.Errorf("%s: unknown config key %q", path, strings.Join(key, "."))
		}
	}
	return fmt.Errorf("%s: %w", path, err)
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
