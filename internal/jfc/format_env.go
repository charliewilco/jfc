package jfc

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func formatDotenv(input []byte, config Config) ([]byte, error) {
	if !utf8.Valid(input) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}

	text := normalizeLineEndingsToLF(string(input))
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		formatted, err := formatDotenvLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		lines[i] = formatted
	}

	return applyOutputConventions(strings.Join(lines, "\n"), config), nil
}

func formatDotenvLine(line string) (string, error) {
	line = strings.TrimRightFunc(line, unicode.IsSpace)
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return line, nil
	}

	prefix := ""
	rest := trimmed
	if strings.HasPrefix(rest, "export ") || rest == "export" {
		if rest == "export" {
			return line, nil
		}
		prefix = "export "
		rest = strings.TrimLeftFunc(rest[len("export "):], unicode.IsSpace)
	}

	eq := strings.IndexByte(rest, '=')
	if eq < 0 {
		return line, nil
	}

	key := strings.TrimSpace(rest[:eq])
	value := strings.TrimSpace(rest[eq+1:])
	if !validDotenvKey(key) {
		return "", fmt.Errorf("invalid dotenv key %q", key)
	}
	return prefix + key + "=" + value, nil
}

func validDotenvKey(key string) bool {
	if key == "" {
		return false
	}
	for i, r := range key {
		switch {
		case r == '_':
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}
