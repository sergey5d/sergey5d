package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var whitespaceRE = regexp.MustCompile(`\s+`)

var voidTags = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"link":   true,
	"meta":   true,
	"param":  true,
	"source": true,
	"track":  true,
	"wbr":    true,
}

type node interface {
	isNode()
}

type textNode struct {
	Text string
}

func (textNode) isNode() {}

type elementNode struct {
	Tag      string
	Attrs    [][2]string
	Children []node
}

func (elementNode) isNode() {}

type parser struct {
	src string
	pos int
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <source-html>\n", os.Args[0])
		os.Exit(2)
	}

	input, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read source: %v\n", err)
		os.Exit(1)
	}

	out, err := ConvertHTMLToBark(string(input))
	if err != nil {
		fmt.Fprintf(os.Stderr, "convert html to bark: %v\n", err)
		os.Exit(1)
	}

	if _, err := io.WriteString(os.Stdout, out); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}

func ConvertHTMLToBark(source string) (string, error) {
	p := &parser{src: source}
	nodes, err := p.parseNodes("")
	if err != nil {
		return "", err
	}

	var lines []string
	for _, n := range nodes {
		lines = append(lines, formatNode(n, 0)...)
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func (p *parser) parseNodes(stopTag string) ([]node, error) {
	var nodes []node
	for !p.eof() {
		if p.startsWith("<!--") {
			p.skipComment()
			continue
		}
		if p.startsWith("<!") {
			p.skipDeclaration()
			continue
		}
		if stopTag != "" && p.startsWith("</") {
			tag, err := p.parseClosingTag()
			if err != nil {
				return nil, err
			}
			if strings.EqualFold(tag, stopTag) {
				return nodes, nil
			}
			continue
		}
		if p.startsWith("<") {
			elem, selfClosing, err := p.parseStartTag()
			if err != nil {
				return nil, err
			}
			if !selfClosing && !voidTags[strings.ToLower(elem.Tag)] {
				if strings.EqualFold(elem.Tag, "script") || strings.EqualFold(elem.Tag, "style") {
					raw, err := p.readRawUntilClosingTag(elem.Tag)
					if err != nil {
						return nil, err
					}
					if raw != "" {
						elem.Children = append(elem.Children, textNode{Text: unescapeHTML(raw)})
					}
				} else {
					children, err := p.parseNodes(elem.Tag)
					if err != nil {
						return nil, err
					}
					elem.Children = children
				}
			}
			nodes = append(nodes, elem)
			continue
		}

		text := p.readUntil("<")
		if text != "" {
			nodes = append(nodes, textNode{Text: unescapeHTML(text)})
		}
	}

	if stopTag != "" {
		return nil, fmt.Errorf("unterminated <%s>", stopTag)
	}
	return nodes, nil
}

func (p *parser) parseStartTag() (elementNode, bool, error) {
	if !p.consume("<") {
		return elementNode{}, false, fmt.Errorf("expected '<' at %d", p.pos)
	}

	tag := p.readName()
	if tag == "" {
		return elementNode{}, false, fmt.Errorf("expected tag name at %d", p.pos)
	}

	var attrs [][2]string
	selfClosing := false

	for !p.eof() {
		p.skipSpaces()
		switch {
		case p.startsWith("/>"):
			p.pos += 2
			selfClosing = true
			return elementNode{Tag: tag, Attrs: attrs}, selfClosing, nil
		case p.startsWith(">"):
			p.pos++
			return elementNode{Tag: tag, Attrs: attrs}, selfClosing, nil
		default:
			key := p.readAttrName()
			if key == "" {
				return elementNode{}, false, fmt.Errorf("expected attribute name in <%s> at %d", tag, p.pos)
			}
			p.skipSpaces()
			value := ""
			if p.consume("=") {
				p.skipSpaces()
				value = p.readAttrValue()
			}
			attrs = append(attrs, [2]string{key, unescapeHTML(value)})
		}
	}

	return elementNode{}, false, fmt.Errorf("unterminated <%s>", tag)
}

func (p *parser) parseClosingTag() (string, error) {
	if !p.consume("</") {
		return "", fmt.Errorf("expected closing tag at %d", p.pos)
	}
	tag := p.readName()
	p.skipSpaces()
	if !p.consume(">") {
		return "", fmt.Errorf("unterminated closing tag </%s>", tag)
	}
	return tag, nil
}

func (p *parser) readRawUntilClosingTag(tag string) (string, error) {
	lower := strings.ToLower(p.src[p.pos:])
	needle := "</" + strings.ToLower(tag) + ">"
	idx := strings.Index(lower, needle)
	if idx < 0 {
		return "", fmt.Errorf("unterminated raw <%s> block", tag)
	}
	raw := p.src[p.pos : p.pos+idx]
	p.pos += idx + len(needle)
	return raw, nil
}

func (p *parser) skipComment() {
	if idx := strings.Index(p.src[p.pos:], "-->"); idx >= 0 {
		p.pos += idx + 3
		return
	}
	p.pos = len(p.src)
}

func (p *parser) skipDeclaration() {
	if idx := strings.Index(p.src[p.pos:], ">"); idx >= 0 {
		p.pos += idx + 1
		return
	}
	p.pos = len(p.src)
}

func (p *parser) readUntil(token string) string {
	if idx := strings.Index(p.src[p.pos:], token); idx >= 0 {
		out := p.src[p.pos : p.pos+idx]
		p.pos += idx
		return out
	}
	out := p.src[p.pos:]
	p.pos = len(p.src)
	return out
}

func (p *parser) readName() string {
	start := p.pos
	for !p.eof() {
		r := rune(p.src[p.pos])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			p.pos++
			continue
		}
		break
	}
	return p.src[start:p.pos]
}

func (p *parser) readAttrName() string {
	start := p.pos
	for !p.eof() {
		r := rune(p.src[p.pos])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == ':' {
			p.pos++
			continue
		}
		break
	}
	return p.src[start:p.pos]
}

func (p *parser) readAttrValue() string {
	if p.eof() {
		return ""
	}
	switch p.src[p.pos] {
	case '"', '\'':
		quote := p.src[p.pos]
		p.pos++
		start := p.pos
		for !p.eof() && p.src[p.pos] != quote {
			p.pos++
		}
		value := p.src[start:p.pos]
		if !p.eof() {
			p.pos++
		}
		return value
	default:
		start := p.pos
		for !p.eof() {
			r := rune(p.src[p.pos])
			if unicode.IsSpace(r) || r == '>' || (r == '/' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '>') {
				break
			}
			p.pos++
		}
		return p.src[start:p.pos]
	}
}

func (p *parser) skipSpaces() {
	for !p.eof() && unicode.IsSpace(rune(p.src[p.pos])) {
		p.pos++
	}
}

func (p *parser) startsWith(prefix string) bool {
	return strings.HasPrefix(p.src[p.pos:], prefix)
}

func (p *parser) consume(prefix string) bool {
	if p.startsWith(prefix) {
		p.pos += len(prefix)
		return true
	}
	return false
}

func (p *parser) eof() bool {
	return p.pos >= len(p.src)
}

func normalizeText(text string) string {
	return strings.TrimSpace(whitespaceRE.ReplaceAllString(text, " "))
}

func formatAttrValue(value string) string {
	if value != "" && regexp.MustCompile(`^[^=\s\[\]"]+$`).MatchString(value) {
		return value
	}
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}

func splitFields(value string) []string {
	return strings.Fields(value)
}

func formatNode(n node, indent int) []string {
	switch typed := n.(type) {
	case textNode:
		text := normalizeText(typed.Text)
		if text == "" {
			return nil
		}
		return []string{strings.Repeat(" ", indent) + text}
	case elementNode:
		return formatElement(typed, indent)
	default:
		return nil
	}
}

func formatElement(elem elementNode, indent int) []string {
	indentStr := strings.Repeat(" ", indent)
	tag := elem.Tag
	if tag == "div" {
		tag = ""
	}

	var ids []string
	var classes []string
	var otherAttrs [][2]string
	for _, attr := range elem.Attrs {
		switch attr[0] {
		case "id":
			ids = splitFields(attr[1])
		case "class":
			classes = splitFields(attr[1])
		default:
			otherAttrs = append(otherAttrs, attr)
		}
	}

	head := "["
	if tag != "" {
		head += tag
	}
	if len(ids) > 0 {
		if head == "[" {
			head += "@" + ids[0]
			ids = ids[1:]
		}
		for _, id := range ids {
			head += " @" + id
		}
	}
	if len(classes) > 0 {
		if head == "[" {
			head += "<: "
		} else {
			head += " <: "
		}
		head += strings.Join(classes, ", ")
	}
	if len(otherAttrs) > 0 {
		parts := make([]string, 0, len(otherAttrs))
		for _, attr := range otherAttrs {
			parts = append(parts, attr[0]+"="+formatAttrValue(attr[1]))
		}
		if head == "[" {
			head += strings.Join(parts, " ")
		} else {
			head += " " + strings.Join(parts, " ")
		}
	}
	hasMetadata := len(ids) > 0 || len(classes) > 0 || len(otherAttrs) > 0

	var children []node
	for _, child := range elem.Children {
		switch c := child.(type) {
		case textNode:
			if normalizeText(c.Text) == "" {
				continue
			}
		}
		children = append(children, child)
	}

	if len(children) == 0 {
		return []string{indentStr + head + "]"}
	}

	onlyText := true
	var textParts []string
	for _, child := range children {
		t, ok := child.(textNode)
		if !ok {
			onlyText = false
			break
		}
		textParts = append(textParts, normalizeText(t.Text))
	}
	if onlyText {
		body := strings.Join(textParts, " ")
		if hasMetadata {
			return []string{fmt.Sprintf("%s%s | %s]", indentStr, head, body)}
		}
		if tag != "" {
			return []string{fmt.Sprintf("%s%s %s]", indentStr, head, body)}
		}
		return []string{fmt.Sprintf("%s%s%s]", indentStr, head, body)}
	}

	lines := []string{indentStr + head}
	for _, child := range children {
		lines = append(lines, formatNode(child, indent+2)...)
	}
	lines = append(lines, indentStr+"]")
	return lines
}

func unescapeHTML(s string) string {
	replacer := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
		"&nbsp;", " ",
	)
	return replacer.Replace(s)
}
