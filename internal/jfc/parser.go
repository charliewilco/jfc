package jfc

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type valueKind int

const (
	kindNull valueKind = iota
	kindBool
	kindNumber
	kindString
	kindArray
	kindObject
)

type value struct {
	kind    valueKind
	boolVal bool
	raw     string
	array   []*value
	object  []member
}

type member struct {
	key   string
	value *value
}

type syntaxError struct {
	line    int
	column  int
	message string
}

func (e *syntaxError) Error() string {
	return fmt.Sprintf("line %d, column %d: %s", e.line, e.column, e.message)
}

type parser struct {
	input  []byte
	pos    int
	line   int
	column int
}

func parseJSON(input []byte) (*value, error) {
	p := &parser{
		input:  input,
		line:   1,
		column: 1,
	}

	p.skipWhitespace()
	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if !p.done() {
		return nil, p.errorf("unexpected trailing content")
	}

	return value, nil
}

func (p *parser) parseValue() (*value, error) {
	if p.done() {
		return nil, p.errorf("unexpected end of input")
	}

	switch p.peek() {
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case '"':
		text, err := p.parseString()
		if err != nil {
			return nil, err
		}
		return &value{kind: kindString, raw: text}, nil
	case 't':
		if err := p.expectKeyword("true"); err != nil {
			return nil, err
		}
		return &value{kind: kindBool, boolVal: true}, nil
	case 'f':
		if err := p.expectKeyword("false"); err != nil {
			return nil, err
		}
		return &value{kind: kindBool, boolVal: false}, nil
	case 'n':
		if err := p.expectKeyword("null"); err != nil {
			return nil, err
		}
		return &value{kind: kindNull}, nil
	default:
		if isNumberStart(p.peek()) {
			number, err := p.parseNumber()
			if err != nil {
				return nil, err
			}
			return &value{kind: kindNumber, raw: number}, nil
		}
		return nil, p.errorf("unexpected character %q", p.peek())
	}
}

func (p *parser) parseObject() (*value, error) {
	p.advance()
	p.skipWhitespace()

	object := &value{kind: kindObject}
	if p.consume('}') {
		return object, nil
	}

	for {
		if p.done() || p.peek() != '"' {
			return nil, p.errorf("expected object key string")
		}

		key, err := p.parseString()
		if err != nil {
			return nil, err
		}

		p.skipWhitespace()
		if !p.consume(':') {
			return nil, p.errorf("expected ':' after object key")
		}

		p.skipWhitespace()
		child, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		object.object = append(object.object, member{key: key, value: child})

		p.skipWhitespace()
		if p.consume('}') {
			return object, nil
		}
		if !p.consume(',') {
			return nil, p.errorf("expected ',' or '}' in object")
		}
		p.skipWhitespace()
	}
}

func (p *parser) parseArray() (*value, error) {
	p.advance()
	p.skipWhitespace()

	array := &value{kind: kindArray}
	if p.consume(']') {
		return array, nil
	}

	for {
		child, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		array.array = append(array.array, child)

		p.skipWhitespace()
		if p.consume(']') {
			return array, nil
		}
		if !p.consume(',') {
			return nil, p.errorf("expected ',' or ']' in array")
		}
		p.skipWhitespace()
	}
}

func (p *parser) parseString() (string, error) {
	start := p.pos
	p.advance()

	for {
		if p.done() {
			return "", p.errorf("unterminated string")
		}

		ch := p.advance()
		switch ch {
		case '"':
			var result string
			if err := json.Unmarshal(p.input[start:p.pos], &result); err != nil {
				return "", p.errorf("invalid string escape sequence")
			}
			return result, nil
		case '\\':
			if p.done() {
				return "", p.errorf("unterminated escape sequence")
			}
			esc := p.advance()
			if esc == 'u' {
				for range 4 {
					if p.done() {
						return "", p.errorf("unterminated unicode escape")
					}
					if !isHexDigit(p.advance()) {
						return "", p.errorf("invalid unicode escape")
					}
				}
			}
		case '\n', '\r':
			return "", p.errorf("unterminated string")
		}
	}
}

func (p *parser) parseNumber() (string, error) {
	start := p.pos

	if p.consume('-') {
		if p.done() {
			return "", p.errorf("unexpected end of number")
		}
	}

	switch {
	case p.consume('0'):
		if !p.done() && isDigit(p.peek()) {
			return "", p.errorf("leading zeroes are not allowed")
		}
	case !p.done() && isDigitOneToNine(p.peek()):
		for !p.done() && isDigit(p.peek()) {
			p.advance()
		}
	default:
		return "", p.errorf("invalid number")
	}

	if p.consume('.') {
		if p.done() || !isDigit(p.peek()) {
			return "", p.errorf("fractional part requires at least one digit")
		}
		for !p.done() && isDigit(p.peek()) {
			p.advance()
		}
	}

	if !p.done() && (p.peek() == 'e' || p.peek() == 'E') {
		p.advance()
		if !p.done() && (p.peek() == '+' || p.peek() == '-') {
			p.advance()
		}
		if p.done() || !isDigit(p.peek()) {
			return "", p.errorf("exponent requires at least one digit")
		}
		for !p.done() && isDigit(p.peek()) {
			p.advance()
		}
	}

	return string(p.input[start:p.pos]), nil
}

func (p *parser) expectKeyword(keyword string) error {
	if p.pos+len(keyword) > len(p.input) {
		return p.errorf("unexpected end of input")
	}
	if !bytes.Equal(p.input[p.pos:p.pos+len(keyword)], []byte(keyword)) {
		return p.errorf("unexpected token")
	}
	for range len(keyword) {
		p.advance()
	}
	return nil
}

func (p *parser) skipWhitespace() {
	for !p.done() {
		switch p.peek() {
		case ' ', '\t', '\n', '\r':
			p.advance()
		default:
			return
		}
	}
}

func (p *parser) consume(target byte) bool {
	if p.done() || p.peek() != target {
		return false
	}
	p.advance()
	return true
}

func (p *parser) peek() byte {
	return p.input[p.pos]
}

func (p *parser) advance() byte {
	ch := p.input[p.pos]
	p.pos++
	if ch == '\n' {
		p.line++
		p.column = 1
		return ch
	}
	p.column++
	return ch
}

func (p *parser) done() bool {
	return p.pos >= len(p.input)
}

func (p *parser) errorf(format string, args ...any) error {
	return &syntaxError{
		line:    p.line,
		column:  p.column,
		message: fmt.Sprintf(format, args...),
	}
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isDigitOneToNine(b byte) bool {
	return b >= '1' && b <= '9'
}

func isNumberStart(b byte) bool {
	return b == '-' || isDigit(b)
}

func isHexDigit(b byte) bool {
	switch {
	case b >= '0' && b <= '9':
		return true
	case b >= 'a' && b <= 'f':
		return true
	case b >= 'A' && b <= 'F':
		return true
	default:
		return false
	}
}
