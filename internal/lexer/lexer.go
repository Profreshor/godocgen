package lexer

import (
	"fmt"
)

type Lexer struct {
	Tokens []Token
	source []byte
	pos    int
	lang   Language
}

// Initialize mutable Lexer
func CreateLexer(content []byte, ext string) (*Lexer, error) {
	lang, supported := SupportedLanguages[ext]
	if !supported {
		return nil, fmt.Errorf("unsupported extension: %s", ext)
	}
	return &Lexer{
		source: content,
		pos:    0,
		Tokens: make([]Token, 0),
		lang:   lang,
	}, nil
}

// Tokenize input file
func (lex *Lexer) Tokenize() {
	var kind Tokenkind = ILLEGAL
	for lex.isValid() {
		start := lex.pos
		switch ch := lex.peek(); {
		case isSpace(ch):
			lex.consumeSpace()
		case isIdent(ch):
			kind = lex.consumeIdent(start)
		case isDigit(ch):
			kind = lex.consumeDigit()
		case ch == '"':
			kind = lex.consumeString()
		case lex.isPunct(ch):
			kind = lex.consumePunct()
		case lex.isOperator(ch):
			kind = lex.consumeOperator()
		default:
			continue
		}
		lex.emitToken(kind, start)
	}

}

func (lex *Lexer) isValid() bool {
	return lex.pos < len(lex.source)
}

func (lex *Lexer) isPunct(char byte) bool {
	if _, exists := lex.lang.Literals[string(char)]; exists {
		return true
	}
	return false
}

func (lex *Lexer) isOperator(char byte) bool {
	if _, exists := lex.lang.Literals[string(char)]; exists {
		return true
	}
	return false
}

func (lex *Lexer) peek() byte {
	if !lex.isValid() {
		return 0
	}
	return lex.source[lex.pos]
}

func (lex *Lexer) consumeSpace() {
	for lex.isValid() && isSpace(lex.peek()) {
		lex.pos++
	}
}
func (lex *Lexer) consumeIdent(start int) Tokenkind {
	for lex.isValid() && (isLetter(lex.peek()) || isDigit(lex.peek())) {
		lex.pos++
	}
	literal := string(lex.source[start:lex.pos])
	if kind, found := lex.lang.Literals[literal]; found {
		return kind
	}
	return IDENT
}

func (lex *Lexer) consumeDigit() Tokenkind {
	for lex.isValid() && isDigit(lex.peek()) {
		lex.pos++
	}
	return NUMBER
}

func (lex *Lexer) consumeString() Tokenkind {
	lex.pos++
	for lex.isValid() {
		ch := lex.peek()
		if ch == '\\' {
			lex.pos += 2
			continue
		}
		if ch == '"' {
			lex.pos++
			return STRING
		}
	}
	return ILLEGAL
}

func (lex *Lexer) consumePunct() Tokenkind {
	for lex.isValid() && lex.isPunct(lex.peek()) {
		lex.pos++
	}
	return PUNCT
}

func (lex *Lexer) consumeOperator() Tokenkind {
	for lex.isValid() && lex.isPunct(lex.peek()) {
		lex.pos++
	}
	return OPERATOR
}

func (lex *Lexer) emitToken(kind Tokenkind, start int) {
	lex.Tokens = append(lex.Tokens, Token{
		Kind: kind,
		Span: Span{
			Start: Pos{Byte: start},
			End:   Pos{Byte: lex.pos},
		},
	})
}
