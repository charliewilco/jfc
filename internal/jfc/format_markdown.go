package jfc

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

func formatMarkdown(input []byte, config Config) ([]byte, error) {
	if !utf8.Valid(input) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}

	_ = goldmark.DefaultParser().Parse(text.NewReader(input))

	lines := strings.Split(normalizeLineEndingsToLF(string(input)), "\n")
	inFence := false
	fenceMarker := ""
	for i, line := range lines {
		if fence, ok := markdownFence(line); ok {
			line = strings.TrimLeftFunc(line, unicode.IsSpace)
			if !inFence {
				inFence = true
				fenceMarker = fence
			} else if strings.HasPrefix(line, fenceMarker) {
				inFence = false
				fenceMarker = ""
			}
			lines[i] = line
			continue
		}

		if !inFence && strings.TrimSpace(line) == "" {
			lines[i] = ""
		}
	}

	return applyOutputConventions(strings.Join(lines, "\n"), config), nil
}

func markdownFence(line string) (string, bool) {
	trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)
	if strings.HasPrefix(trimmed, "```") {
		return "```", true
	}
	if strings.HasPrefix(trimmed, "~~~") {
		return "~~~", true
	}
	return "", false
}
