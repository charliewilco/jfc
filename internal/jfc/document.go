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
	FormatXML      FormatKind = "xml"
	FormatCSV      FormatKind = "csv"
	FormatTSV      FormatKind = "tsv"
	FormatDotenv   FormatKind = "dotenv"
	FormatHCL      FormatKind = "hcl"
)

func detectFormat(path string) (FormatKind, bool) {
	base := strings.ToLower(filepath.Base(path))
	ext := strings.ToLower(filepath.Ext(path))
	if base == ".env" || strings.HasPrefix(base, ".env.") || ext == ".env" {
		return FormatDotenv, true
	}

	switch ext {
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
	case ".xml", ".svg", ".plist", ".xib", ".storyboard", ".csproj", ".vbproj", ".fsproj", ".props", ".targets", ".pom":
		return FormatXML, true
	case ".csv":
		return FormatCSV, true
	case ".tsv":
		return FormatTSV, true
	case ".hcl", ".tf", ".tfvars", ".nomad":
		return FormatHCL, true
	default:
		return "", false
	}
}

func supportedExtensionsText() string {
	return ".json, .jsonc, .jsonl, .ndjson, .yaml, .yml, .toml, .md, .markdown, .xml, .svg, .plist, .xib, .storyboard, .csproj, .vbproj, .fsproj, .props, .targets, .pom, .csv, .tsv, .env, .env.*, *.env, .hcl, .tf, .tfvars, .nomad"
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
	case FormatXML:
		return formatXML(input, config)
	case FormatCSV:
		return formatCSV(input, config)
	case FormatTSV:
		return formatTSV(input, config)
	case FormatDotenv:
		return formatDotenv(input, config)
	case FormatHCL:
		return formatHCL(input, config)
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}

func applyFinalNewlineOnly(input []byte, config Config) []byte {
	if len(input) == 0 {
		if config.TrailingNewline {
			return []byte(config.endOfLineString())
		}
		return nil
	}

	output := append([]byte(nil), input...)
	for len(output) > 0 && (output[len(output)-1] == '\n' || output[len(output)-1] == '\r') {
		output = output[:len(output)-1]
	}
	if config.TrailingNewline {
		output = append(output, []byte(config.endOfLineString())...)
	}
	return output
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
