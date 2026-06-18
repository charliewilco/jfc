package jfc

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

func formatJSONL(input []byte, config Config) ([]byte, error) {
	scanner := bufio.NewScanner(bytes.NewReader(input))
	scanner.Buffer(make([]byte, 1024), 1024*1024*64)

	lineNumber := 0
	var lines []string
	lineConfig := config
	lineConfig.ArrayExpand = ExpandNever
	lineConfig.ObjectExpand = ExpandNever
	lineConfig.TrailingNewline = false
	lineConfig.EndOfLine = EndOfLineLF

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		formatted, err := formatJSON([]byte(line), lineConfig)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		lines = append(lines, string(formatted))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read JSONL: %w", err)
	}

	return applyOutputConventions(strings.Join(lines, "\n"), config), nil
}
