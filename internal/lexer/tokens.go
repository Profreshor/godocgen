package lexer

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
		return "punctuation"
	case OPERATOR:
		return "operator"
	default:
		return "unknown"
	}
}

// Handle whitespaces
func isSpace(char byte) bool {
	if char == ' ' || char == '\t' || char == '\n' || char == '\r' {
		return true
	}
	return false
}

// Handle Identifiers
func isIdent(char byte) bool {
	if isLetter(char) {
		return true
	}
	return false
}

func isLetter(char byte) bool {
	if ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_' {
		return true
	}
	return false
}

func isDigit(char byte) bool {
	if char >= '0' && char <= '9' {
		return true
	}
	return false
}
