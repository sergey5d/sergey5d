package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
)

// Braketo syntax:
//
//   [ body ]                                   -> <div>body</div>
//   [p body]                                   -> <p>body</p>
//   [h1 @intro <: hero, title {data-mode="demo"} Hello]
//     -> <h1 id="intro" class="hero title" data-mode="demo">Hello</h1>
//   [p @id1 @id2 <: body {attr1=value attr2="string value"} text]
//   [a href=/docs target=_blank | Read more]
//
// Rules:
//   - A leading "[" starts an element.
//   - If the next rune is alphanumeric, it becomes the tag name.
//   - Otherwise the tag defaults to div.
//   - @name appends an id token, so "@root @hero" becomes id="root hero".
//   - <: class1, class2 stores class names; commas separate classes.
//     Because class names cannot contain spaces, the class list ends once the
//     parser sees non-comma-separated text or another metadata block.
//   - {...} stores arbitrary attributes as key=value pairs.
//   - Bare key=value attributes are also allowed before a "|" body separator.
//   - Nested [...] blocks are parsed recursively.

type parser struct {
	src []rune
	pos int
}

type node struct {
	Tag     string
	Classes []string
	IDs     []string
	Attrs   map[string]string
	Content []content
}

type content struct {
	Text string
	Node *node
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <source-file>\n", os.Args[0])
		os.Exit(2)
	}

	input, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read source: %v\n", err)
		os.Exit(1)
	}

	htmlOut, err := ParseBraketo(string(input))
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse source: %v\n", err)
		os.Exit(1)
	}

	if _, err := io.WriteString(os.Stdout, htmlOut); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}

func ParseBraketo(input string) (string, error) {
	p := &parser{src: []rune(input)}
	var nodes []*node
	var textParts []string

	for !p.eof() {
		if p.peek() == '[' {
			n, err := p.parseNode()
			if err != nil {
				return "", err
			}
			nodes = append(nodes, n)
			continue
		}

		text := p.readUntil('[')
		if strings.TrimSpace(text) != "" {
			textParts = append(textParts, text)
		}
	}

	var out strings.Builder
	hasExplicitHTMLRoot := len(nodes) > 0 && strings.EqualFold(nodes[0].Tag, "html")
	if !hasExplicitHTMLRoot {
		out.WriteString("<html>")
	}

	for _, text := range textParts {
		out.WriteString(escapeHTML(text))
	}
	for _, n := range nodes {
		out.WriteString(n.HTML())
	}

	if !hasExplicitHTMLRoot {
		out.WriteString("</html>")
	}

	return out.String(), nil
}

func (n *node) HTML() string {
	var b strings.Builder
	b.WriteByte('<')
	b.WriteString(n.Tag)

	if len(n.Classes) > 0 {
		b.WriteString(` class="`)
		b.WriteString(escapeHTML(strings.Join(n.Classes, " ")))
		b.WriteByte('"')
	}

	if len(n.IDs) > 0 {
		b.WriteString(` id="`)
		b.WriteString(escapeHTML(strings.Join(n.IDs, " ")))
		b.WriteByte('"')
	}

	if len(n.Attrs) > 0 {
		keys := make([]string, 0, len(n.Attrs))
		for k := range n.Attrs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteByte(' ')
			b.WriteString(k)
			b.WriteString(`="`)
			b.WriteString(escapeHTML(n.Attrs[k]))
			b.WriteByte('"')
		}
	}

	b.WriteByte('>')
	for _, item := range n.Content {
		if item.Node != nil {
			b.WriteString(item.Node.HTML())
		} else {
			b.WriteString(escapeHTML(item.Text))
		}
	}
	b.WriteString("</")
	b.WriteString(n.Tag)
	b.WriteByte('>')
	return b.String()
}

func (p *parser) parseNode() (*node, error) {
	if err := p.expect('['); err != nil {
		return nil, err
	}

	n := &node{
		Tag:   "div",
		Attrs: map[string]string{},
	}

	if r := p.peek(); isNameStart(r) {
		n.Tag = p.readName()
	}

	for {
		p.skipSpaces()
		switch p.peek() {
		case '@':
			id, err := p.parseIDShortcut()
			if err != nil {
				return nil, err
			}
			n.IDs = append(n.IDs, id)
		case '<':
			items, err := p.parseClassList()
			if err != nil {
				return nil, err
			}
			n.Classes = items
		case '{':
			attrs, err := p.parsePropsBlock()
			if err != nil {
				return nil, err
			}
			for k, v := range attrs {
				n.Attrs[k] = v
			}
		default:
			if p.looksLikeBareAttr() {
				key, value, err := p.parseBareAttr()
				if err != nil {
					return nil, err
				}
				n.Attrs[key] = value
				continue
			}
			goto body
		}
	}

body:
	p.skipSpaces()
	if p.peek() == '|' {
		p.pos++
		if !p.eof() && unicode.IsSpace(p.peek()) {
			p.skipSpaces()
		}
	}

	var text bytes.Buffer
	flushText := func() {
		if text.Len() == 0 {
			return
		}
		n.Content = append(n.Content, content{Text: text.String()})
		text.Reset()
	}

	for {
		if p.eof() {
			return nil, fmt.Errorf("unterminated element <%s>", n.Tag)
		}

		switch p.peek() {
		case '[':
			flushText()
			child, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, content{Node: child})
		case ']':
			p.pos++
			flushText()
			return n, nil
		default:
			text.WriteRune(p.next())
		}
	}
}

func (p *parser) parseClassList() ([]string, error) {
	if err := p.expect('<'); err != nil {
		return nil, err
	}
	if err := p.expect(':'); err != nil {
		return nil, err
	}

	var classes []string
	p.skipSpaces()

	for !p.eof() {
		start := p.pos
		for !p.eof() && isClassNamePart(p.peek()) {
			p.pos++
		}
		if start == p.pos {
			break
		}
		classes = append(classes, string(p.src[start:p.pos]))

		checkpoint := p.pos
		p.skipSpaces()
		if !p.eof() && p.peek() == ',' {
			p.pos++
			p.skipSpaces()
			continue
		}
		p.pos = checkpoint
		break
	}

	return classes, nil
}

func (p *parser) looksLikeBareAttr() bool {
	if p.eof() {
		return false
	}
	i := p.pos
	if !isAttrNamePart(p.src[i]) {
		return false
	}
	for i < len(p.src) && isAttrNamePart(p.src[i]) {
		i++
	}
	return i < len(p.src) && p.src[i] == '='
}

func (p *parser) parseBareAttr() (string, string, error) {
	start := p.pos
	for !p.eof() && isAttrNamePart(p.peek()) {
		p.pos++
	}
	if start == p.pos {
		return "", "", fmt.Errorf("expected attribute name at rune %d", p.pos)
	}
	key := string(p.src[start:p.pos])

	if err := p.expect('='); err != nil {
		return "", "", err
	}

	if p.eof() {
		return "", "", fmt.Errorf("expected attribute value for %q at rune %d", key, p.pos)
	}

	var value string
	switch p.peek() {
	case '"', '\'':
		quote := p.next()
		start = p.pos
		for !p.eof() && p.peek() != quote {
			p.pos++
		}
		if p.eof() {
			return "", "", fmt.Errorf("unterminated quoted value for %q", key)
		}
		value = string(p.src[start:p.pos])
		p.pos++
	default:
		start = p.pos
		for !p.eof() {
			r := p.peek()
			if unicode.IsSpace(r) || r == '|' || r == '[' || r == ']' {
				break
			}
			p.pos++
		}
		value = string(p.src[start:p.pos])
	}

	return key, value, nil
}

func (p *parser) parsePropsBlock() (map[string]string, error) {
	if err := p.expect('{'); err != nil {
		return nil, err
	}
	raw := p.readUntilMatching('}')
	if raw == "" && p.eof() {
		return nil, fmt.Errorf("unterminated {...} block")
	}

	attrs := map[string]string{}

	for _, assignment := range splitAssignments(raw) {
		parts := strings.SplitN(assignment, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		key = strings.Trim(key, `"`)
		key = strings.Trim(key, `'`)
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"`)
		value = strings.Trim(value, `'`)
		attrs[key] = value
	}
	return attrs, nil
}

func splitCSVRespectingQuotes(raw string) []string {
	var out []string
	var cur strings.Builder
	inQuote := rune(0)

	flush := func() {
		part := strings.TrimSpace(cur.String())
		part = strings.Trim(part, `"`)
		part = strings.Trim(part, `'`)
		if part != "" {
			out = append(out, part)
		}
		cur.Reset()
	}

	for _, r := range raw {
		if inQuote != 0 {
			if r == inQuote {
				inQuote = 0
			} else {
				cur.WriteRune(r)
			}
			continue
		}
		switch r {
		case '"', '\'':
			inQuote = r
		case ',':
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return out
}

func splitAssignments(raw string) []string {
	var out []string
	var cur strings.Builder
	inQuote := rune(0)
	seenEquals := false

	flush := func() {
		part := strings.TrimSpace(cur.String())
		if part != "" {
			out = append(out, part)
		}
		cur.Reset()
		seenEquals = false
	}

	for _, r := range raw {
		if inQuote != 0 {
			cur.WriteRune(r)
			if r == inQuote {
				inQuote = 0
			}
			continue
		}
		switch r {
		case '"', '\'':
			inQuote = r
			cur.WriteRune(r)
		case '=':
			seenEquals = true
			cur.WriteRune(r)
		case ' ', '\n', '\t', '\r':
			if seenEquals {
				flush()
			} else {
				cur.WriteRune(r)
			}
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return out
}

func (p *parser) parseIDShortcut() (string, error) {
	if err := p.expect('@'); err != nil {
		return "", err
	}
	start := p.pos
	for !p.eof() {
		r := p.peek()
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			p.pos++
			continue
		}
		break
	}
	if start == p.pos {
		return "", fmt.Errorf("expected id after @ at rune %d", p.pos)
	}
	return string(p.src[start:p.pos]), nil
}

func (p *parser) readUntilMatching(close rune) string {
	start := p.pos
	inQuote := rune(0)
	for !p.eof() {
		r := p.next()
		if inQuote != 0 {
			if r == inQuote {
				inQuote = 0
			}
			continue
		}
		if r == '"' || r == '\'' {
			inQuote = r
			continue
		}
		if r == close {
			return string(p.src[start : p.pos-1])
		}
	}
	return ""
}

func (p *parser) readUntil(stop rune) string {
	start := p.pos
	for !p.eof() && p.peek() != stop {
		p.pos++
	}
	return string(p.src[start:p.pos])
}

func (p *parser) readName() string {
	start := p.pos
	for !p.eof() && isNamePart(p.peek()) {
		p.pos++
	}
	return string(p.src[start:p.pos])
}

func (p *parser) skipSpaces() {
	for !p.eof() && unicode.IsSpace(p.peek()) {
		p.pos++
	}
}

func (p *parser) expect(want rune) error {
	if p.eof() || p.peek() != want {
		return fmt.Errorf("expected %q at rune %d", want, p.pos)
	}
	p.pos++
	return nil
}

func (p *parser) eof() bool {
	return p.pos >= len(p.src)
}

func (p *parser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) next() rune {
	r := p.peek()
	p.pos++
	return r
}

func isNameStart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isNamePart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_'
}

func isClassNamePart(r rune) bool {
	return isNamePart(r)
}

func isAttrNamePart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == ':'
}

func escapeHTML(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
