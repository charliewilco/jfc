package jfc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type xmlDocument struct {
	children []*xmlNode
}

type xmlNode struct {
	kind        xmlNodeKind
	name        xml.Name
	attr        []xml.Attr
	text        string
	comment     string
	target      string
	instruction string
	directive   string
	children    []*xmlNode
}

type xmlNodeKind int

const (
	xmlNodeElement xmlNodeKind = iota
	xmlNodeComment
	xmlNodeProcInst
	xmlNodeDirective
)

func formatXML(input []byte, config Config) ([]byte, error) {
	if !utf8.Valid(input) {
		return nil, fmt.Errorf("input is not valid UTF-8")
	}

	if containsXMLCDATA(input) {
		if err := validateXML(input); err != nil {
			return nil, err
		}
		return applyOutputConventions(string(input), config), nil
	}

	document, preserveOnly, err := parseXMLDocument(input)
	if err != nil {
		return nil, err
	}
	if preserveOnly {
		return applyOutputConventions(string(input), config), nil
	}

	output := renderXMLDocument(document, config)
	return applyOutputConventions(output, config), nil
}

func containsXMLCDATA(input []byte) bool {
	return bytes.Contains(input, []byte("<![CDATA["))
}

func validateXML(input []byte) error {
	_, _, err := parseXMLDocument(input)
	return err
}

func parseXMLDocument(input []byte) (xmlDocument, bool, error) {
	decoder := xml.NewDecoder(bytes.NewReader(input))
	document := xmlDocument{}
	stack := []*xmlNode{}
	rootElements := 0
	preserveOnly := false

	for {
		token, err := decoder.RawToken()
		if err != nil {
			if err == io.EOF {
				break
			}
			return xmlDocument{}, false, err
		}

		switch tok := token.(type) {
		case xml.StartElement:
			for _, attr := range tok.Attr {
				if xmlAttrRequiresPreserve(attr) {
					preserveOnly = true
				}
			}
			node := &xmlNode{
				kind: xmlNodeElement,
				name: tok.Name,
				attr: append([]xml.Attr(nil), tok.Attr...),
			}
			if len(stack) == 0 {
				rootElements++
				if rootElements > 1 {
					return xmlDocument{}, false, fmt.Errorf("XML syntax error: multiple root elements")
				}
				document.children = append(document.children, node)
			} else {
				parent := stack[len(stack)-1]
				if parent.text != "" {
					preserveOnly = true
				}
				parent.children = append(parent.children, node)
			}
			stack = append(stack, node)
		case xml.EndElement:
			if len(stack) == 0 {
				return xmlDocument{}, false, fmt.Errorf("XML syntax error: unexpected end element </%s>", xmlRenderName(tok.Name))
			}
			current := stack[len(stack)-1]
			if current.name.Space != tok.Name.Space || current.name.Local != tok.Name.Local {
				return xmlDocument{}, false, fmt.Errorf("XML syntax error: expected </%s>, got </%s>", xmlRenderName(current.name), xmlRenderName(tok.Name))
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			text := string(tok)
			if len(stack) == 0 {
				if strings.TrimSpace(text) != "" {
					return xmlDocument{}, false, fmt.Errorf("XML syntax error: character data outside root element")
				}
				continue
			}
			current := stack[len(stack)-1]
			if strings.TrimSpace(text) == "" {
				if text != "" && !strings.ContainsAny(text, "\r\n") && len(current.children) == 0 {
					preserveOnly = true
				}
				continue
			}
			if len(current.children) > 0 {
				preserveOnly = true
			}
			if strings.ContainsAny(text, "\r\n\t") {
				preserveOnly = true
			}
			current.text += text
		case xml.Comment:
			node := &xmlNode{kind: xmlNodeComment, comment: string(tok)}
			if len(stack) == 0 {
				document.children = append(document.children, node)
			} else {
				parent := stack[len(stack)-1]
				if parent.text != "" {
					preserveOnly = true
				}
				parent.children = append(parent.children, node)
			}
		case xml.ProcInst:
			node := &xmlNode{kind: xmlNodeProcInst, target: tok.Target, instruction: string(tok.Inst)}
			if len(stack) == 0 {
				document.children = append(document.children, node)
			} else {
				parent := stack[len(stack)-1]
				if parent.text != "" {
					preserveOnly = true
				}
				parent.children = append(parent.children, node)
			}
		case xml.Directive:
			node := &xmlNode{kind: xmlNodeDirective, directive: string(tok)}
			if len(stack) == 0 {
				document.children = append(document.children, node)
			} else {
				parent := stack[len(stack)-1]
				if parent.text != "" {
					preserveOnly = true
				}
				parent.children = append(parent.children, node)
			}
		}
	}

	if len(stack) > 0 {
		current := stack[len(stack)-1]
		return xmlDocument{}, false, fmt.Errorf("XML syntax error: missing end element </%s>", xmlRenderName(current.name))
	}
	if rootElements == 0 {
		return xmlDocument{}, false, fmt.Errorf("XML syntax error: missing root element")
	}
	return document, preserveOnly, nil
}

func xmlAttrRequiresPreserve(attr xml.Attr) bool {
	if attr.Name.Space == "xml" && attr.Name.Local == "space" && attr.Value == "preserve" {
		return true
	}
	return strings.ContainsAny(attr.Value, "\r\n\t")
}

func renderXMLDocument(document xmlDocument, config Config) string {
	lines := make([]string, 0, len(document.children))
	for _, child := range document.children {
		renderXMLNode(child, 0, config, &lines)
	}
	return strings.Join(lines, "\n")
}

func renderXMLNode(node *xmlNode, depth int, config Config, lines *[]string) {
	indent := xmlIndent(depth, config)
	switch node.kind {
	case xmlNodeElement:
		*lines = append(*lines, renderXMLElement(node, depth, config)...)
	case xmlNodeComment:
		*lines = append(*lines, indent+"<!--"+node.comment+"-->")
	case xmlNodeProcInst:
		if strings.TrimSpace(node.instruction) == "" {
			*lines = append(*lines, indent+"<?"+node.target+"?>")
		} else {
			*lines = append(*lines, indent+"<?"+node.target+" "+node.instruction+"?>")
		}
	case xmlNodeDirective:
		*lines = append(*lines, indent+"<!"+node.directive+">")
	}
}

func renderXMLElement(node *xmlNode, depth int, config Config) []string {
	indent := xmlIndent(depth, config)
	start := indent + "<" + xmlRenderName(node.name) + xmlRenderAttrs(node.attr)
	if len(node.children) == 0 && node.text == "" {
		return []string{start + "/>"}
	}
	if len(node.children) == 0 {
		return []string{start + ">" + xmlEscapeText(node.text) + "</" + xmlRenderName(node.name) + ">"}
	}

	lines := []string{start + ">"}
	for _, child := range node.children {
		renderXMLNode(child, depth+1, config, &lines)
	}
	lines = append(lines, indent+"</"+xmlRenderName(node.name)+">")
	return lines
}

func xmlRenderName(name xml.Name) string {
	if name.Space == "" {
		return name.Local
	}
	return name.Space + ":" + name.Local
}

func xmlRenderAttrs(attrs []xml.Attr) string {
	if len(attrs) == 0 {
		return ""
	}

	parts := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		parts = append(parts, xmlRenderName(attr.Name)+"=\""+xmlEscapeAttr(attr.Value)+"\"")
	}
	return " " + strings.Join(parts, " ")
}

func xmlIndent(depth int, config Config) string {
	if depth <= 0 {
		return ""
	}
	if config.UseTabs {
		return strings.Repeat("\t", depth)
	}
	return strings.Repeat(" ", depth*config.TabWidth)
}

func xmlEscapeText(value string) string {
	var builder strings.Builder
	for _, r := range value {
		switch r {
		case '&':
			builder.WriteString("&amp;")
		case '<':
			builder.WriteString("&lt;")
		case '>':
			builder.WriteString("&gt;")
		default:
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func xmlEscapeAttr(value string) string {
	var builder strings.Builder
	for _, r := range value {
		switch r {
		case '&':
			builder.WriteString("&amp;")
		case '<':
			builder.WriteString("&lt;")
		case '>':
			builder.WriteString("&gt;")
		case '"':
			builder.WriteString("&quot;")
		case '\t':
			builder.WriteString("&#x9;")
		case '\n':
			builder.WriteString("&#xA;")
		case '\r':
			builder.WriteString("&#xD;")
		default:
			if unicode.IsControl(r) {
				builder.WriteString(fmt.Sprintf("&#x%X;", r))
				continue
			}
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
