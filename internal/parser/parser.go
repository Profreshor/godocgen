package parser

import (
	"strings"

	"github.com/Profreshor/godocgen/internal/lexer"
)

type Parser struct {
	tokens  []lexer.Token
	source  []byte
	pos     int
	symbols []Symbol
}

func CreateParser(tokens []lexer.Token, src []byte) *Parser {
	return &Parser{
		tokens:  tokens,
		source:  src,
		pos:     0,
		symbols: make([]Symbol, 0),
	}
}

func (p *Parser) Parse() {
	for p.isValid() {
		start := p.pos
		currentTok := p.peek()
		if !p.isKeyword(currentTok.Kind) {
			p.advance()
			continue
		}
		keyword := p.sliceSpan(currentTok)
		p.advance()
		switch keyword {
		case "func":
			p.parseFunc(start)
		case "type":
			p.parseValueSpec(start, TYPE)
		case "const":
			p.parseValueSpec(start, CONSTANT)
		case "var":
			p.parseValueSpec(start, VARIABLE)
		case "package":
			p.parsePackage(start)
		case "import":
			p.parseImport(start)
		}
	}
}

// Positional Helper Methods
func (p *Parser) peek() lexer.Token {
	if !p.isValid() {
		return lexer.Token{Kind: lexer.EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) isValid() bool {
	if p.pos < len(p.tokens) {
		if p.tokens[p.pos].Kind != lexer.EOF {
			return true
		}
		return false
	}
	return false
}

func (p *Parser) isKeyword(kind lexer.Tokenkind) bool {
	return kind == lexer.KEYWORD
}

func (p *Parser) skipBalanced(opener, closer string) bool {
	if !p.expectText(opener) {
		return false
	}
	if !p.scanTo(closer) {
		return false
	}
	p.advance()
	return true
}

func (p *Parser) scanTo(target string) bool {
	depth := 0
	for p.isValid() {
		text := p.sliceSpan(p.peek())
		if depth == 0 && text == target {
			return true
		}
		switch text {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			depth--
		}
		p.advance()
	}
	return false
}

func (p *Parser) newlineCount(i int) int {
	if i <= 0 || i >= len(p.tokens) {
		return 2
	}
	gap := p.source[p.tokens[i-1].Span.End.Byte:p.tokens[i].Span.Start.Byte]
	return strings.Count(string(gap), "\n")
}

func (p *Parser) newlineBefore(i int) bool {
	return p.newlineCount(i) > 0
}

func (p *Parser) scanToLineEnd() {
	depth := 0
	for p.isValid() {
		if depth == 0 && p.newlineBefore(p.pos) {
			return
		}
		switch p.sliceSpan(p.peek()) {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			depth--
		}
		if depth < 0 {
			return
		}
		p.advance()
	}
}

func (p *Parser) expect(kind lexer.Tokenkind) (lexer.Token, bool) {
	tok := p.peek()
	if tok.Kind != kind {
		return lexer.Token{}, false
	}
	p.advance()
	return tok, true
}

func (p *Parser) expectText(text string) bool {
	if p.sliceSpan(p.peek()) != text {
		return false
	}
	p.advance()
	return true
}

// Token Parsing Methods
func (p *Parser) parseOneSpec(start int, kind SymbolKind) bool {
	var names []lexer.Token
	for {
		tok, ok := p.expect(lexer.IDENT)
		if !ok {
			break
		}
		names = append(names, tok)
		if !p.expectText(",") {
			break
		}
	}
	if kind == TYPE {
		if p.sliceSpan(p.peek()) == "[" {
			p.skipBalanced("[", "]")
		}
		switch p.sliceSpan(p.peek()) {
		case "struct":
			kind = STRUCT
		case "interface":
			kind = INTERFACE
		}
	}

	p.scanToLineEnd()
	if len(names) == 0 {
		return false
	}
	span := p.spanFrom(start)
	detail := p.textOf(span)
	doc := p.collectDoc(start)
	for _, name := range names {
		p.emitSymbol(kind, p.sliceSpan(name), span, name.Span, doc, detail, "")
	}
	return true
}

func (p *Parser) parseValueSpec(start int, kind SymbolKind) bool {
	if p.sliceSpan(p.peek()) == "(" {
		p.advance()
		for p.isValid() && p.sliceSpan(p.peek()) != ")" {
			specStart := p.pos
			p.parseOneSpec(specStart, kind)
			if p.pos == specStart {
				p.advance()
			}
		}
		p.expectText(")")
		return true
	}
	return p.parseOneSpec(start, kind)
}

func (p *Parser) advance() {
	p.pos++
}

func (p *Parser) parseFunc(start int) bool {
	var owner string = ""
	if p.sliceSpan(p.peek()) == "(" {
		ownerStr, ok := p.readReceiver()
		if ok {
			owner = ownerStr
		}
	}
	name, ok := p.expect(lexer.IDENT)
	if !ok {
		return false
	}
	if !p.scanTo("{") {
		return false
	}
	sigStart := p.tokens[start].Span.Start.Byte
	sigEnd := p.peek().Span.Start.Byte
	detail := strings.TrimSpace(string(p.source[sigStart:sigEnd]))
	kind := FUNCTION
	doc := p.collectDoc(start)
	if owner != "" {
		kind = METHOD
	}
	p.skipBalanced("{", "}")
	p.emitSymbol(kind, p.sliceSpan(name), p.spanFrom(start), name.Span, doc, detail, owner)
	return true
}

func (p *Parser) readReceiver() (string, bool) {
	p.advance()
	var idents []lexer.Token
	for p.isValid() && p.sliceSpan(p.peek()) != ")" && p.sliceSpan(p.peek()) != "[" {
		if p.peek().Kind == lexer.IDENT {
			idents = append(idents, p.peek())
		}
		p.advance()
	}
	if !p.scanTo(")") {
		return "", false
	}
	p.advance()
	if len(idents) == 0 {
		return "", false
	}
	return p.sliceSpan(idents[len(idents)-1]), true
}

func (p *Parser) parsePackage(start int) bool {
	tok, ok := p.expect(lexer.IDENT)
	if !ok {
		return false
	}
	doc := p.collectDoc(start)
	p.emitSymbol(MODULE, p.sliceSpan(tok), p.spanFrom(start), tok.Span, doc, p.textOf(tok.Span), "")
	return true
}

func (p *Parser) parseImport(start int) bool {
	if p.sliceSpan(p.peek()) == "(" {
		p.advance()
		for p.isValid() && p.sliceSpan(p.peek()) != ")" {
			before := p.pos
			p.parseOneImport(p.pos)
			if p.pos == before {
				p.advance()
			}
		}
		p.expectText(")")
		return true
	}
	return p.parseOneImport(start)
}

func (p *Parser) parseOneImport(start int) bool {
	alias, hasAlias := p.expect(lexer.IDENT)
	p.expectText(".")

	path, ok := p.expect(lexer.STRING)
	if !ok {
		p.scanToLineEnd()
		return false
	}
	name := strings.Trim(p.sliceSpan(path), "`\"")
	selector := path.Span
	detail := p.textOf(selector)
	doc := p.collectDoc(start)
	if hasAlias {
		name = p.sliceSpan(alias)
		selector = alias.Span
	}
	p.emitSymbol(PACKAGE, name, p.spanFrom(start), selector, doc, detail, "")
	return true
}

func (p *Parser) collectDoc(start int) string {
	first := start
	i := start - 1
	for i >= 0 && p.isValid() {
		if p.tokens[i].Kind != lexer.COMMENT {
			break
		}
		if p.newlineCount(i+1) == 2 {
			break
		}
		if p.newlineCount(i) == 0 {
			break
		}
		first = i
		i = i - 1
	}
	if first == start {
		return ""
	}
	return string(p.source[p.tokens[first].Span.Start.Byte:p.tokens[start-1].Span.End.Byte])
}

// Reader Helper Functions
func (p *Parser) sliceSpan(tok lexer.Token) string {
	return string(p.source[tok.Span.Start.Byte:tok.Span.End.Byte])
}

func (p *Parser) spanFrom(start int) lexer.Span {
	return lexer.Span{
		Start: p.tokens[start].Span.Start,
		End:   p.tokens[p.pos-1].Span.End,
	}
}

func (p *Parser) textOf(span lexer.Span) string {
	return strings.TrimSpace(string(p.source[span.Start.Byte:span.End.Byte]))
}

// Token emitting helper
func (p *Parser) emitSymbol(symbol SymbolKind, name string, span lexer.Span, selector lexer.Span, doc string, detail string, owner string) {
	p.symbols = append(p.symbols, Symbol{
		Kind:     symbol,
		Name:     name,
		Span:     span,
		Selector: selector,
		Doc:      doc,
		Detail:   detail,
		Owner:    owner,
	})
}

// Symbols Accessor
func (p *Parser) Symbols() []Symbol {
	return p.symbols
}
