package lexer

import (
	"fmt"
	"unicode/utf8"
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
		r, width := lex.peekRune()
		switch {
		case isSpace(r):
			lex.pos += width
			continue
		case isIdent(r):
			kind = lex.consumeIdent(start)
		case isDigit(r):
			kind = lex.consumeDigit()
		case r == '`':
			kind = lex.consumeDelimited('`', false)
		case r == '"':
			kind = lex.consumeDelimited('"', true)
		case r == '\'':
			kind = lex.consumeDelimited('\'', true)
		case lex.isPunctOrOper(r):
			kind = lex.consumePunctOrOper()
		default:
			lex.pos += width
		}
		lex.emitToken(kind, start)
	}
	lex.emitToken(EOF, lex.pos)
}

func (lex *Lexer) isValid() bool {
	return lex.pos < len(lex.source)
}

func (lex *Lexer) isPunctOrOper(r rune) bool {
	if kind, exists := lex.lang.Literals[string(r)]; exists {
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

func (lex *Lexer) peekRune() (rune, int) {
	if !lex.isValid() {
		return 0, 0
	}
	return utf8.DecodeRune(lex.source[lex.pos:])
}

func (lex *Lexer) consumeIdent(start int) Tokenkind {
	for lex.isValid() {
		r, width := lex.peekRune()
		if !(isLetter(r) || isDigit(r)) {
			break
		}
		lex.pos += width
	}
	literal := string(lex.source[start:lex.pos])
	if kind, found := lex.lang.Literals[literal]; found {
		return kind
	}
	return IDENT
}

func (lex *Lexer) consumeDigit() Tokenkind {
	for (lex.isValid() && isDigit(rune(lex.peek()))) || lex.peek() == '.' {
		lex.pos++
	}
	return NUMBER
}

func (lex *Lexer) consumeDelimited(closer byte, allowEscapes bool) Tokenkind {
	lex.pos++
	for lex.isValid() {
		ch := lex.peek()
		if allowEscapes {
			if ch == '\\' {
				lex.pos++
				if lex.isValid() {
					lex.pos++
				}
				continue
			}
		}
		if ch == closer {
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
		if kind, exists := lex.lang.Literals[twoChars]; exists && (kind == PUNCT || kind == OPERATOR) {
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
