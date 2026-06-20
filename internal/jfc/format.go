package jfc

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

func formatJSON(input []byte, config Config) ([]byte, error) {
	if !utf8.Valid(input) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}

	root, err := parseJSON(input)
	if err != nil {
		return nil, err
	}

	renderer := formatter{config: config}
	output := renderer.renderValue(root, 0)

	return applyOutputConventions(output, config), nil
}

type formatter struct {
	config Config
}

func (f formatter) renderValue(node *value, depth int) string {
	switch node.kind {
	case kindNull:
		return "null"
	case kindBool:
		if node.boolVal {
			return "true"
		}
		return "false"
	case kindNumber:
		return node.raw
	case kindString:
		return quoteJSONString(node.raw)
	case kindArray:
		return f.renderArray(node.array, depth)
	case kindObject:
		return f.renderObject(node.object, depth)
	default:
		return ""
	}
}

func (f formatter) renderArray(items []*value, depth int) string {
	if len(items) == 0 {
		if f.config.SpaceWithinBrackets {
			return "[ ]"
		}
		return "[]"
	}

	if f.config.ArrayExpand != ExpandAlways {
		if inline, ok := f.renderInlineArray(items); ok && (f.config.ArrayExpand == ExpandNever || f.fitsWidth(inline, depth)) {
			return inline
		}
	}

	lines := make([]string, 0, len(items)+2)
	lines = append(lines, "[")
	for i, item := range items {
		line := f.indent(depth+1) + f.renderValue(item, depth+1)
		if i < len(items)-1 {
			line += ","
		}
		lines = append(lines, line)
	}
	lines = append(lines, f.indent(depth)+"]")
	return strings.Join(lines, "\n")
}

func (f formatter) renderObject(items []member, depth int) string {
	if len(items) == 0 {
		if f.config.SpaceWithinBraces {
			return "{ }"
		}
		return "{}"
	}

	items = f.orderedMembers(items)
	if f.config.ObjectExpand != ExpandAlways {
		if inline, ok := f.renderInlineObject(items); ok && (f.config.ObjectExpand == ExpandNever || f.fitsWidth(inline, depth)) {
			return inline
		}
	}

	lines := make([]string, 0, len(items)+2)
	lines = append(lines, "{")
	for i, item := range items {
		line := f.indent(depth+1) + quoteJSONString(item.key) + f.colonSpacing() + f.renderValue(item.value, depth+1)
		if i < len(items)-1 {
			line += ","
		}
		lines = append(lines, line)
	}
	lines = append(lines, f.indent(depth)+"}")
	return strings.Join(lines, "\n")
}

func (f formatter) renderInlineValue(node *value) (string, bool) {
	switch node.kind {
	case kindNull, kindBool, kindNumber, kindString:
		return f.renderValue(node, 0), true
	case kindArray:
		if f.config.ArrayExpand == ExpandAlways {
			return "", false
		}
		return f.renderInlineArray(node.array)
	case kindObject:
		if f.config.ObjectExpand == ExpandAlways {
			return "", false
		}
		return f.renderInlineObject(f.orderedMembers(node.object))
	default:
		return "", false
	}
}

func (f formatter) renderInlineArray(items []*value) (string, bool) {
	if len(items) == 0 {
		if f.config.SpaceWithinBrackets {
			return "[ ]", true
		}
		return "[]", true
	}

	parts := make([]string, 0, len(items))
	for _, item := range items {
		rendered, ok := f.renderInlineValue(item)
		if !ok {
			return "", false
		}
		parts = append(parts, rendered)
	}

	inside := strings.Join(parts, ", ")
	if f.config.SpaceWithinBrackets {
		return "[ " + inside + " ]", true
	}
	return "[" + inside + "]", true
}

func (f formatter) renderInlineObject(items []member) (string, bool) {
	if len(items) == 0 {
		if f.config.SpaceWithinBraces {
			return "{ }", true
		}
		return "{}", true
	}

	parts := make([]string, 0, len(items))
	for _, item := range items {
		rendered, ok := f.renderInlineValue(item.value)
		if !ok {
			return "", false
		}
		parts = append(parts, quoteJSONString(item.key)+f.colonSpacing()+rendered)
	}

	inside := strings.Join(parts, ", ")
	if f.config.SpaceWithinBraces {
		return "{ " + inside + " }", true
	}
	return "{" + inside + "}", true
}

func (f formatter) orderedMembers(items []member) []member {
	if !f.config.SortKeys {
		return items
	}

	sorted := slices.Clone(items)
	slices.SortFunc(sorted, func(a member, b member) int {
		switch {
		case a.key < b.key:
			return -1
		case a.key > b.key:
			return 1
		default:
			return 0
		}
	})
	return sorted
}

func (f formatter) fitsWidth(rendered string, depth int) bool {
	return f.indentWidth(depth)+displayWidth(rendered, f.config.TabWidth) <= f.config.PrintWidth
}

func (f formatter) indent(depth int) string {
	return strings.Repeat(f.config.indentUnit(), depth)
}

func (f formatter) indentWidth(depth int) int {
	if f.config.UseTabs {
		return depth * f.config.TabWidth
	}
	return depth * len(f.config.indentUnit())
}

func (f formatter) colonSpacing() string {
	if f.config.SpaceAfterColon {
		return ": "
	}
	return ":"
}

func displayWidth(text string, tabWidth int) int {
	width := 0
	for _, r := range text {
		if r == '\t' {
			width += tabWidth
			continue
		}
		width++
	}
	return width
}

func quoteJSONString(text string) string {
	const hex = "0123456789abcdef"

	var builder strings.Builder
	builder.Grow(len(text) + 2)
	builder.WriteByte('"')
	for _, r := range text {
		switch r {
		case '\\', '"':
			builder.WriteByte('\\')
			builder.WriteRune(r)
		case '\b':
			builder.WriteString(`\b`)
		case '\f':
			builder.WriteString(`\f`)
		case '\n':
			builder.WriteString(`\n`)
		case '\r':
			builder.WriteString(`\r`)
		case '\t':
			builder.WriteString(`\t`)
		case '\u2028':
			builder.WriteString(`\u2028`)
		case '\u2029':
			builder.WriteString(`\u2029`)
		default:
			if r < 0x20 {
				builder.WriteString(`\u00`)
				builder.WriteByte(hex[byte(r)>>4])
				builder.WriteByte(hex[byte(r)&0x0f])
				continue
			}
			builder.WriteRune(r)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}
