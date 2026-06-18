package jfc

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	toml "github.com/pelletier/go-toml/v2"
)

func formatTOML(input []byte, config Config) ([]byte, error) {
	if !utf8.Valid(input) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}
	var decoded any
	if err := toml.Unmarshal(input, &decoded); err != nil {
		return nil, err
	}

	text := normalizeLineEndingsToLF(string(input))
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = formatTOMLLine(line)
	}

	return applyOutputConventions(strings.Join(lines, "\n"), config), nil
}

func formatTOMLLine(line string) string {
	trimmedRight := strings.TrimRightFunc(line, unicode.IsSpace)
	eq := findTOMLEquals(trimmedRight)
	if eq < 0 {
		return trimmedRight
	}

	before := strings.TrimRightFunc(trimmedRight[:eq], unicode.IsSpace)
	after := strings.TrimLeftFunc(trimmedRight[eq+1:], unicode.IsSpace)
	if before == "" || after == "" {
		return trimmedRight
	}
	return before + " = " + after
}

func findTOMLEquals(line string) int {
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
			return -1
		case !inBasicString && !inLiteralString && r == '=':
			return i
		}
	}

	return -1
}
