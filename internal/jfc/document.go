package jfc

import (
	"fmt"
	"path/filepath"
	"strings"
)

type FormatKind string

const (
	FormatJSON     FormatKind = "json"
	FormatJSONC    FormatKind = "jsonc"
	FormatJSONL    FormatKind = "jsonl"
	FormatYAML     FormatKind = "yaml"
	FormatTOML     FormatKind = "toml"
	FormatMarkdown FormatKind = "markdown"
)

func detectFormat(path string) (FormatKind, bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return FormatJSON, true
	case ".jsonc":
		return FormatJSONC, true
	case ".jsonl", ".ndjson":
		return FormatJSONL, true
	case ".yaml", ".yml":
		return FormatYAML, true
	case ".toml":
		return FormatTOML, true
	case ".md", ".markdown":
		return FormatMarkdown, true
	default:
		return "", false
	}
}

func supportedExtensionsText() string {
	return ".json, .jsonc, .jsonl, .ndjson, .yaml, .yml, .toml, .md, .markdown"
}

func formatDocument(input []byte, format FormatKind, config Config) ([]byte, error) {
	switch format {
	case FormatJSON:
		return formatJSON(input, config)
	case FormatJSONC:
		return formatJSONC(input, config)
	case FormatJSONL:
		return formatJSONL(input, config)
	case FormatYAML:
		return formatYAML(input, config)
	case FormatTOML:
		return formatTOML(input, config)
	case FormatMarkdown:
		return formatMarkdown(input, config)
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}

func applyOutputConventions(output string, config Config) []byte {
	output = normalizeLineEndingsToLF(output)
	output = strings.TrimRight(output, "\n")
	if config.TrailingNewline {
		output += "\n"
	}
	if eol := config.endOfLineString(); eol != "\n" {
		output = strings.ReplaceAll(output, "\n", eol)
	}
	return []byte(output)
}

func normalizeLineEndingsToLF(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\r", "\n")
}
