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
	for lex.isValid() {
		kind := ILLEGAL
		start := lex.pos
		switch ch := lex.peek(); {
		case isSpace(ch):
			lex.pos++
			continue
		case isIdent(ch):
			kind = lex.consumeIdent(start)
		case isDigit(ch):
			kind = lex.consumeDigit()
		case ch == '`':
			kind = lex.consumeBacktick()
		case ch == '"':
			kind = lex.consumeString()
		case ch == '\'':
			kind = lex.consumeSingleQuote()
		case lex.isPunctOrOper(ch):
			kind = lex.consumePunctOrOper()
		default:
			lex.pos++
		}
		lex.emitToken(kind, start)
	}

}

func (lex *Lexer) isValid() bool {
	return lex.pos < len(lex.source)
}

func (lex *Lexer) isPunctOrOper(char byte) bool {
	if kind, exists := lex.lang.Literals[string(char)]; exists {
		if kind == PUNCT || kind == OPERATOR {
			return true
		}
		return false
	}
	return false
}

func (lex *Lexer) peek() byte {
	if !lex.isValid() {
		return 0
	}
	return lex.source[lex.pos]
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
		lex.pos++
	}
	return ILLEGAL
}

func (lex *Lexer) consumeSingleQuote() Tokenkind {
	lex.pos++
	for lex.isValid() {
		ch := lex.peek()
		if ch == '\\' {
			lex.pos += 2
			continue
		}
		if ch == '\'' {
			lex.pos++
			return STRING
		}
		lex.pos++
	}
	return ILLEGAL
}

func (lex *Lexer) consumeBacktick() Tokenkind {
	lex.pos++
	for lex.isValid() {
		ch := lex.peek()
		if ch == '\\' {
			lex.pos += 2
			continue
		}
		if ch == '`' {
			lex.pos++
			return STRING
		}
		lex.pos++
	}
	return ILLEGAL
}

func (lex *Lexer) consumePunctOrOper() Tokenkind {
	remaining := len(lex.source) - lex.pos
	if remaining >= 3 {
		threeChars := lex.source[lex.pos : lex.pos+3]
		if kind, exists := lex.lang.Literals[string(threeChars)]; exists && (kind == PUNCT || kind == OPERATOR) {
			lex.pos += 3
			return kind
		}
	}
	if remaining >= 2 {
		twoChars := string(lex.source[lex.pos : lex.pos+2])
		if twoChars == "//" {
			lex.pos += 2
			for lex.isValid() && lex.source[lex.pos] != '\n' {
				lex.pos++
			}
			return COMMENT
		}
		if twoChars == "/*" {
			lex.pos += 2
			for lex.isValid() {
				if len(lex.source)-lex.pos >= 2 && string(lex.source[lex.pos:lex.pos+2]) == "*/" {
					lex.pos += 2
					return COMMENT
				}
				lex.pos++
			}
		}
		if kind, exists := lex.lang.Literals[twoChars]; exists {
			lex.pos += 2
			return kind
		}
	}
	if remaining >= 1 {
		oneChar := lex.source[lex.pos : lex.pos+1]
		if kind, exists := lex.lang.Literals[string(oneChar)]; exists {
			lex.pos += 1
			return kind
		}
	}
	if remaining > 0 {
		lex.pos++
	}
	return ILLEGAL
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
