package jfc

import (
	"fmt"
	"strings"
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
	fence := markdownFence{}
	for i, line := range lines {
		if candidate, ok := markdownFenceSequence(line); ok && !inFence && markdownOpeningFence(candidate) {
			inFence = true
			fence = candidate
			lines[i] = line
			continue
		}

		if candidate, ok := markdownFenceSequence(line); ok && inFence && markdownClosingFence(candidate, fence) {
			inFence = false
			fence = markdownFence{}
			lines[i] = line
			continue
		}

		if !inFence && strings.TrimSpace(line) == "" {
			lines[i] = ""
		}
	}

	return applyOutputConventions(strings.Join(lines, "\n"), config), nil
}

type markdownFence struct {
	marker byte
	length int
	rest   string
}

func markdownFenceSequence(line string) (markdownFence, bool) {
	indent := 0
	for indent < len(line) && line[indent] == ' ' {
		indent++
	}
	if indent > 3 {
		return markdownFence{}, false
	}

	trimmed := line[indent:]
	if len(trimmed) < 3 {
		return markdownFence{}, false
	}

	marker := trimmed[0]
	if marker != '`' && marker != '~' {
		return markdownFence{}, false
	}

	length := 0
	for length < len(trimmed) && trimmed[length] == marker {
		length++
	}
	if length < 3 {
		return markdownFence{}, false
	}

	return markdownFence{
		marker: marker,
		length: length,
		rest:   trimmed[length:],
	}, true
}

func markdownOpeningFence(fence markdownFence) bool {
	return fence.marker != '`' || !strings.Contains(fence.rest, "`")
}

func markdownClosingFence(candidate markdownFence, opening markdownFence) bool {
	return candidate.marker == opening.marker &&
		candidate.length >= opening.length &&
		strings.TrimSpace(candidate.rest) == ""
}
