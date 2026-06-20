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
	state := tomlMultilineNone
	for i, line := range lines {
		if state == tomlMultilineNone {
			lines[i] = formatTOMLLine(line)
		}
		state = nextTOMLMultilineState(line, state)
	}

	return applyOutputConventions(strings.Join(lines, "\n"), config), nil
}

type tomlMultilineState int

const (
	tomlMultilineNone tomlMultilineState = iota
	tomlMultilineBasic
	tomlMultilineLiteral
)

func formatTOMLLine(line string) string {
	preserveTrailing := nextTOMLMultilineState(line, tomlMultilineNone) != tomlMultilineNone
	candidate := line
	if !preserveTrailing {
		candidate = strings.TrimRightFunc(line, unicode.IsSpace)
	}

	eq := findTOMLEquals(candidate)
	if eq < 0 {
		return candidate
	}

	before := strings.TrimRightFunc(candidate[:eq], unicode.IsSpace)
	after := strings.TrimLeftFunc(candidate[eq+1:], unicode.IsSpace)
	if before == "" || after == "" {
		return candidate
	}
	return before + " = " + after
}

func nextTOMLMultilineState(line string, state tomlMultilineState) tomlMultilineState {
	inBasicString := false
	inLiteralString := false
	escaped := false

	for i := 0; i < len(line); i++ {
		switch state {
		case tomlMultilineBasic:
			if strings.HasPrefix(line[i:], `"""`) && !hasOddBackslashRun(line[:i]) {
				state = tomlMultilineNone
				i += 2
			}
			continue
		case tomlMultilineLiteral:
			if strings.HasPrefix(line[i:], `'''`) {
				state = tomlMultilineNone
				i += 2
			}
			continue
		}

		ch := line[i]
		switch {
		case escaped:
			escaped = false
		case inBasicString && ch == '\\':
			escaped = true
		case inBasicString && ch == '"':
			inBasicString = false
		case inLiteralString && ch == '\'':
			inLiteralString = false
		case inBasicString || inLiteralString:
		case ch == '#':
			return state
		case strings.HasPrefix(line[i:], `"""`):
			state = tomlMultilineBasic
			i += 2
		case strings.HasPrefix(line[i:], `'''`):
			state = tomlMultilineLiteral
			i += 2
		case ch == '"':
			inBasicString = true
		case ch == '\'':
			inLiteralString = true
		}
	}

	return state
}

func hasOddBackslashRun(prefix string) bool {
	count := 0
	for i := len(prefix) - 1; i >= 0 && prefix[i] == '\\'; i-- {
		count++
	}
	return count%2 == 1
}

func findTOMLEquals(line string) int {
	var (
		inBasicString   bool
		inLiteralString bool
		escaped         bool
		nesting         int
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
		case !inBasicString && !inLiteralString && (r == '[' || r == '{'):
			nesting++
		case !inBasicString && !inLiteralString && (r == ']' || r == '}') && nesting > 0:
			nesting--
		case !inBasicString && !inLiteralString && nesting == 0 && r == '=':
			return i
		}
	}

	return -1
}
