package lexer

import "unicode"

type Pos struct {
	Byte int
}

type Span struct {
	Start Pos
	End   Pos
}

type Token struct {
	Kind Tokenkind
	Span Span
}

type Tokenkind uint8

const (
	ILLEGAL Tokenkind = iota
	EOF
	KEYWORD
	IDENT
	STRING
	NUMBER
	COMMENT
	PUNCT
	OPERATOR
)

func (k Tokenkind) String() string {
	switch k {
	case ILLEGAL:
		return "illegal"
	case EOF:
		return "eof"
	case KEYWORD:
		return "keyword"
	case IDENT:
		return "ident"
	case STRING:
		return "string"
	case NUMBER:
		return "number"
	case COMMENT:
		return "comment"
	case PUNCT:
		return "punct"
	case OPERATOR:
		return "operator"
	default:
		return "unknown"
	}
}

// Handle whitespaces
func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// Handle Identifiers
func isIdent(r rune) bool {
	return isLetter(r)
}

func isLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}
