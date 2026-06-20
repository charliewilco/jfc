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

	text := normalizeLineEndingsToLF(string(input))
	lines := strings.Split(text, "\n")
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

		if !inFence && markdownBlankLineCanNormalize(line) {
			lines[i] = ""
		}
	}

	output := strings.Join(lines, "\n")
	if inFence {
		if eol := config.endOfLineString(); eol != "\n" {
			output = strings.ReplaceAll(output, "\n", eol)
		}
		return []byte(output), nil
	}
	return applyOutputConventions(output, config), nil
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

func markdownBlankLineCanNormalize(line string) bool {
	indentWidth := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case ' ':
			indentWidth++
		case '\t':
			return false
		default:
			return false
		}
	}
	return indentWidth < 4
}
