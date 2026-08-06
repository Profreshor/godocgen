package lexer

import (
	"bytes"
	"fmt"
	"sort"
	"unicode/utf8"
)

type Lexer struct {
	Tokens     []Token
	source     []byte
	pos        int
	lang       Language
	lineStarts []int
}

func NewLexer(content []byte, lang Language) *Lexer {
	starts := []int{0}
	for i := range content {
		if content[i] == '\n' {
			starts = append(starts, i+1)
		}
	}
	return &Lexer{
		source:     content,
		pos:        0,
		Tokens:     make([]Token, 0),
		lang:       lang,
		lineStarts: starts,
	}
}

func CreateLexer(content []byte, ext string) (*Lexer, error) {
	lang, supported := supportedLanguages[ext]
	if !supported {
		return nil, fmt.Errorf("unsupported extension: %s", ext)
	}
	return NewLexer(content, lang), nil
}

// Tokenize input file
func (lex *Lexer) Tokenize() {
	for lex.isValid() {
		kind := ILLEGAL
		start := lex.pos
		r, width := lex.peekRune()
		str, isString := lex.stringSyntax(r)
		switch {
		case isSpace(r):
			lex.pos += width
			continue
		case isIdent(r):
			kind = lex.consumeIdent(start)
		case isDigit(r):
			kind = lex.consumeDigit()
		case lex.startsComment():
			kind = lex.consumeComment()
		case isString:
			kind = lex.consumeDelimited(str.Opener, str.Escapes)
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

func (lex *Lexer) Position(offset int) (int, int) {
	i := sort.SearchInts(lex.lineStarts, offset+1) - 1
	line := i + 1
	col := offset - lex.lineStarts[i] + 1
	return line, col
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

func (lex *Lexer) stringSyntax(r rune) (StringSyntax, bool) {
	for _, s := range lex.lang.Strings {
		if rune(s.Opener) == r {
			return s, true
		}
	}
	return StringSyntax{}, false
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

func (lex *Lexer) startsComment() bool {
	rest := lex.source[lex.pos:]
	if lex.lang.LineComment != "" && bytes.HasPrefix(rest, []byte(lex.lang.LineComment)) {
		return true
	}
	if lex.lang.BlockComment.Open != "" && bytes.HasPrefix(rest, []byte(lex.lang.BlockComment.Open)) {
		return true
	}
	return false
}

func (lex *Lexer) consumeComment() Tokenkind {
	rest := lex.source[lex.pos:]
	if lex.lang.LineComment != "" && bytes.HasPrefix(rest, []byte(lex.lang.LineComment)) {
		lex.pos += len(lex.lang.LineComment)
		for lex.isValid() && lex.source[lex.pos] != '\n' {
			lex.pos++
		}
		return COMMENT
	}
	lex.pos += len(lex.lang.BlockComment.Open)
	closer := []byte(lex.lang.BlockComment.Close)
	for lex.isValid() {
		if bytes.HasPrefix(lex.source[lex.pos:], closer) {
			lex.pos += len(closer)
			return COMMENT
		}
		lex.pos++
	}
	return ILLEGAL
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
