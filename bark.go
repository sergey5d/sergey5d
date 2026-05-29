package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var barkWhitespaceRE = regexp.MustCompile(`\s+`)
var barkAttrValueRE = regexp.MustCompile(`^[^=\s\[\]"]+$`)

var barkVoidTags = map[string]bool{
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

func main() {
	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s [gen|import|degen|-g|-i|-d] <file-or-pattern>\n", os.Args[0])
		os.Exit(2)
	}

	mode := "gen"
	pattern := ""
	if len(os.Args) == 2 {
		pattern = os.Args[1]
	} else {
		mode = os.Args[1]
		pattern = os.Args[2]
	}

	var err error
	switch mode {
	case "gen", "-g":
		err = barkGenerateHTML(pattern)
	case "import", "degen", "-i":
		err = barkReverseGenerate(pattern)
	case "-d":
		err = barkReverseGenerate(pattern)
	default:
		err = fmt.Errorf("unknown mode %q, expected gen, import, degen, -g, -i, or -d", mode)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "bark: %v\n", err)
		os.Exit(1)
	}
}

func barkGenerateHTML(pattern string) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("pattern %q matched no files", pattern)
	}

	for _, src := range files {
		input, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read %s: %w", src, err)
		}
		htmlOut, err := ParseBark(string(input))
		if err != nil {
			return fmt.Errorf("parse %s: %w", src, err)
		}
		out := strings.TrimSuffix(src, filepath.Ext(src)) + ".html"
		if err := os.WriteFile(out, []byte(htmlOut), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", out, err)
		}
	}

	return nil
}

func barkReverseGenerate(pattern string) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("pattern %q matched no files", pattern)
	}

	for _, src := range files {
		input, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read %s: %w", src, err)
		}
		barkOut, err := ConvertHTMLToBarkGo(string(input))
		if err != nil {
			return fmt.Errorf("convert %s: %w", src, err)
		}
		out := strings.TrimSuffix(filepath.Base(src), filepath.Ext(src)) + ".bark"
		if err := os.WriteFile(out, []byte(barkOut), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", out, err)
		}
	}

	return nil
}

// Bark -> HTML

type barkParser struct {
	src []rune
	pos int
}

type barkNode struct {
	Tag     string
	Classes []string
	IDs     []string
	Attrs   map[string]string
	Content []barkContent
}

type barkContent struct {
	Text string
	Node *barkNode
}

func ParseBark(input string) (string, error) {
	p := &barkParser{src: []rune(input)}
	var nodes []*barkNode
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
		out.WriteString(barkEscapeHTML(text))
	}
	for _, n := range nodes {
		out.WriteString(n.HTML())
	}

	if !hasExplicitHTMLRoot {
		out.WriteString("</html>")
	}

	return out.String(), nil
}

func (n *barkNode) HTML() string {
	var b strings.Builder
	b.WriteByte('<')
	b.WriteString(n.Tag)

	if len(n.Classes) > 0 {
		b.WriteString(` class="`)
		b.WriteString(barkEscapeHTML(strings.Join(n.Classes, " ")))
		b.WriteByte('"')
	}

	if len(n.IDs) > 0 {
		b.WriteString(` id="`)
		b.WriteString(barkEscapeHTML(strings.Join(n.IDs, " ")))
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
			b.WriteString(barkEscapeHTML(n.Attrs[k]))
			b.WriteByte('"')
		}
	}

	b.WriteByte('>')
	for _, item := range n.Content {
		if item.Node != nil {
			b.WriteString(item.Node.HTML())
		} else {
			b.WriteString(barkEscapeHTML(item.Text))
		}
	}
	b.WriteString("</")
	b.WriteString(n.Tag)
	b.WriteByte('>')
	return b.String()
}

func (p *barkParser) parseNode() (*barkNode, error) {
	if err := p.expect('['); err != nil {
		return nil, err
	}

	n := &barkNode{
		Tag:   "div",
		Attrs: map[string]string{},
	}

	if r := p.peek(); barkIsNameStart(r) {
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
		default:
			if p.peek() == '{' {
				return nil, fmt.Errorf("curly-brace attribute blocks are no longer supported at rune %d", p.pos)
			}
			if p.startsEscapedBodyLiteral() {
				goto body
			}
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
		n.Content = append(n.Content, barkContent{Text: text.String()})
		text.Reset()
	}

	for {
		if p.eof() {
			return nil, fmt.Errorf("unterminated element <%s>", n.Tag)
		}

		if p.startsEscapedBodyLiteral() {
			p.pos++
			text.WriteRune(p.next())
			continue
		}

		switch p.peek() {
		case '[':
			flushText()
			child, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, barkContent{Node: child})
		case ']':
			p.pos++
			flushText()
			return n, nil
		default:
			text.WriteRune(p.next())
		}
	}
}

func (p *barkParser) parseClassList() ([]string, error) {
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
		for !p.eof() && barkIsClassNamePart(p.peek()) {
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

func (p *barkParser) looksLikeBareAttr() bool {
	if p.eof() {
		return false
	}
	i := p.pos
	if !barkIsAttrNamePart(p.src[i]) {
		return false
	}
	for i < len(p.src) && barkIsAttrNamePart(p.src[i]) {
		i++
	}
	return i < len(p.src) && p.src[i] == '='
}

func (p *barkParser) parseBareAttr() (string, string, error) {
	start := p.pos
	for !p.eof() && barkIsAttrNamePart(p.peek()) {
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

func (p *barkParser) startsEscapedBodyLiteral() bool {
	if p.pos+1 >= len(p.src) || p.src[p.pos] != '\\' {
		return false
	}
	switch p.src[p.pos+1] {
	case '<', '|', '[', '=':
		return true
	default:
		return false
	}
}

func (p *barkParser) parseIDShortcut() (string, error) {
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

func (p *barkParser) readUntil(stop rune) string {
	start := p.pos
	for !p.eof() && p.peek() != stop {
		p.pos++
	}
	return string(p.src[start:p.pos])
}

func (p *barkParser) readName() string {
	start := p.pos
	for !p.eof() && barkIsNamePart(p.peek()) {
		p.pos++
	}
	return string(p.src[start:p.pos])
}

func (p *barkParser) skipSpaces() {
	for !p.eof() && unicode.IsSpace(p.peek()) {
		p.pos++
	}
}

func (p *barkParser) expect(want rune) error {
	if p.eof() || p.peek() != want {
		return fmt.Errorf("expected %q at rune %d", want, p.pos)
	}
	p.pos++
	return nil
}

func (p *barkParser) eof() bool {
	return p.pos >= len(p.src)
}

func (p *barkParser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *barkParser) next() rune {
	r := p.peek()
	p.pos++
	return r
}

func barkIsNameStart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func barkIsNamePart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_'
}

func barkIsClassNamePart(r rune) bool {
	return barkIsNamePart(r)
}

func barkIsAttrNamePart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == ':'
}

func barkEscapeHTML(s string) string {
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

// HTML -> Bark

type htmlNode interface {
	isHTMLNode()
}

type htmlTextNode struct {
	Text string
}

func (htmlTextNode) isHTMLNode() {}

type htmlElementNode struct {
	Tag      string
	Attrs    [][2]string
	Children []htmlNode
}

func (htmlElementNode) isHTMLNode() {}

type htmlParser struct {
	src string
	pos int
}

func ConvertHTMLToBarkGo(source string) (string, error) {
	p := &htmlParser{src: source}
	nodes, err := p.parseNodes("")
	if err != nil {
		return "", err
	}

	var lines []string
	for _, n := range nodes {
		lines = append(lines, formatHTMLNodeAsBark(n, 0)...)
	}
	return strings.Join(lines, "\n") + "\n", nil
}

func (p *htmlParser) parseNodes(stopTag string) ([]htmlNode, error) {
	var nodes []htmlNode
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
			if !selfClosing && !barkVoidTags[strings.ToLower(elem.Tag)] {
				if strings.EqualFold(elem.Tag, "script") || strings.EqualFold(elem.Tag, "style") {
					raw, err := p.readRawUntilClosingTag(elem.Tag)
					if err != nil {
						return nil, err
					}
					if raw != "" {
						elem.Children = append(elem.Children, htmlTextNode{Text: barkUnescapeHTML(raw)})
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
			nodes = append(nodes, htmlTextNode{Text: barkUnescapeHTML(text)})
		}
	}

	if stopTag != "" {
		return nil, fmt.Errorf("unterminated <%s>", stopTag)
	}
	return nodes, nil
}

func (p *htmlParser) parseStartTag() (htmlElementNode, bool, error) {
	if !p.consume("<") {
		return htmlElementNode{}, false, fmt.Errorf("expected '<' at %d", p.pos)
	}
	tag := p.readName()
	if tag == "" {
		return htmlElementNode{}, false, fmt.Errorf("expected tag name at %d", p.pos)
	}

	var attrs [][2]string
	selfClosing := false
	for !p.eof() {
		p.skipSpaces()
		switch {
		case p.startsWith("/>"):
			p.pos += 2
			selfClosing = true
			return htmlElementNode{Tag: tag, Attrs: attrs}, selfClosing, nil
		case p.startsWith(">"):
			p.pos++
			return htmlElementNode{Tag: tag, Attrs: attrs}, selfClosing, nil
		default:
			key := p.readAttrName()
			if key == "" {
				return htmlElementNode{}, false, fmt.Errorf("expected attribute name in <%s> at %d", tag, p.pos)
			}
			p.skipSpaces()
			value := ""
			if p.consume("=") {
				p.skipSpaces()
				value = p.readAttrValue()
			}
			attrs = append(attrs, [2]string{key, barkUnescapeHTML(value)})
		}
	}
	return htmlElementNode{}, false, fmt.Errorf("unterminated <%s>", tag)
}

func (p *htmlParser) parseClosingTag() (string, error) {
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

func (p *htmlParser) readUntil(token string) string {
	if idx := strings.Index(p.src[p.pos:], token); idx >= 0 {
		out := p.src[p.pos : p.pos+idx]
		p.pos += idx
		return out
	}
	out := p.src[p.pos:]
	p.pos = len(p.src)
	return out
}

func (p *htmlParser) readName() string {
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

func (p *htmlParser) readRawUntilClosingTag(tag string) (string, error) {
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

func (p *htmlParser) skipComment() {
	if idx := strings.Index(p.src[p.pos:], "-->"); idx >= 0 {
		p.pos += idx + 3
		return
	}
	p.pos = len(p.src)
}

func (p *htmlParser) skipDeclaration() {
	if idx := strings.Index(p.src[p.pos:], ">"); idx >= 0 {
		p.pos += idx + 1
		return
	}
	p.pos = len(p.src)
}

func (p *htmlParser) readAttrName() string {
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

func (p *htmlParser) readAttrValue() string {
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

func (p *htmlParser) skipSpaces() {
	for !p.eof() && unicode.IsSpace(rune(p.src[p.pos])) {
		p.pos++
	}
}

func (p *htmlParser) startsWith(prefix string) bool {
	return strings.HasPrefix(p.src[p.pos:], prefix)
}

func (p *htmlParser) consume(prefix string) bool {
	if p.startsWith(prefix) {
		p.pos += len(prefix)
		return true
	}
	return false
}

func (p *htmlParser) eof() bool {
	return p.pos >= len(p.src)
}

func normalizeBarkText(text string) string {
	return strings.TrimSpace(barkWhitespaceRE.ReplaceAllString(text, " "))
}

func formatBarkAttrValue(value string) string {
	if value != "" && barkAttrValueRE.MatchString(value) {
		return value
	}
	return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
}

func barkSplitFields(value string) []string {
	return strings.Fields(value)
}

func formatHTMLNodeAsBark(n htmlNode, indent int) []string {
	switch typed := n.(type) {
	case htmlTextNode:
		text := normalizeBarkText(typed.Text)
		if text == "" {
			return nil
		}
		return []string{strings.Repeat(" ", indent) + text}
	case htmlElementNode:
		return formatHTMLElementAsBark(typed, indent)
	default:
		return nil
	}
}

func formatHTMLElementAsBark(elem htmlElementNode, indent int) []string {
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
			ids = barkSplitFields(attr[1])
		case "class":
			classes = barkSplitFields(attr[1])
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
			parts = append(parts, attr[0]+"="+formatBarkAttrValue(attr[1]))
		}
		if head == "[" {
			head += strings.Join(parts, " ")
		} else {
			head += " " + strings.Join(parts, " ")
		}
	}
	hasMetadata := len(ids) > 0 || len(classes) > 0 || len(otherAttrs) > 0

	var children []htmlNode
	for _, child := range elem.Children {
		if t, ok := child.(htmlTextNode); ok {
			if normalizeBarkText(t.Text) == "" {
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
		t, ok := child.(htmlTextNode)
		if !ok {
			onlyText = false
			break
		}
		textParts = append(textParts, normalizeBarkText(t.Text))
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
		lines = append(lines, formatHTMLNodeAsBark(child, indent+2)...)
	}
	lines = append(lines, indentStr+"]")
	return lines
}

func barkUnescapeHTML(s string) string {
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
