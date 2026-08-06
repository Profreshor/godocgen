package lexer

var supportedLanguages = map[string]Language{
	".go": goLang,
}

type Language struct {
	Name         string
	LineComment  string
	BlockComment BlockSyntax
	Strings      []StringSyntax
	Literals     map[string]Tokenkind
}

type StringSyntax struct {
	Opener  byte
	Escapes bool
}

type BlockSyntax struct {
	Open  string
	Close string
}

// Language definitions are read-only after init, never mutated
var goLang = Language{
	Name:         "Go",
	LineComment:  "//",
	BlockComment: BlockSyntax{Open: "/*", Close: "*/"},
	Strings: []StringSyntax{
		{Opener: '`', Escapes: false},
		{Opener: '"', Escapes: true},
		{Opener: '\'', Escapes: true},
	},
	Literals: map[string]Tokenkind{
		"break":       KEYWORD,
		"case":        KEYWORD,
		"chan":        KEYWORD,
		"const":       KEYWORD,
		"continue":    KEYWORD,
		"default":     KEYWORD,
		"defer":       KEYWORD,
		"else":        KEYWORD,
		"fallthrough": KEYWORD,
		"for":         KEYWORD,
		"func":        KEYWORD,
		"go":          KEYWORD,
		"goto":        KEYWORD,
		"if":          KEYWORD,
		"import":      KEYWORD,
		"interface":   KEYWORD,
		"map":         KEYWORD,
		"package":     KEYWORD,
		"range":       KEYWORD,
		"return":      KEYWORD,
		"select":      KEYWORD,
		"struct":      KEYWORD,
		"switch":      KEYWORD,
		"type":        KEYWORD,
		"var":         KEYWORD,
		"(":           PUNCT,
		"[":           PUNCT,
		"{":           PUNCT,
		",":           PUNCT,
		".":           PUNCT,
		")":           PUNCT,
		"]":           PUNCT,
		"}":           PUNCT,
		";":           PUNCT,
		":":           PUNCT,
		"+":           OPERATOR,
		"-":           OPERATOR,
		"*":           OPERATOR,
		"/":           OPERATOR,
		"%":           OPERATOR,
		"&":           OPERATOR,
		"|":           OPERATOR,
		"^":           OPERATOR,
		"<<":          OPERATOR,
		">>":          OPERATOR,
		"&^":          OPERATOR,
		"+=":          OPERATOR,
		"-=":          OPERATOR,
		"*=":          OPERATOR,
		"/=":          OPERATOR,
		"%=":          OPERATOR,
		"&=":          OPERATOR,
		"|=":          OPERATOR,
		"^=":          OPERATOR,
		"<<=":         OPERATOR,
		">>=":         OPERATOR,
		"&^=":         OPERATOR,
		"&&":          OPERATOR,
		"||":          OPERATOR,
		"<-":          OPERATOR,
		"++":          OPERATOR,
		"--":          OPERATOR,
		"==":          OPERATOR,
		"<":           OPERATOR,
		">":           OPERATOR,
		"=":           OPERATOR,
		"!":           OPERATOR,
		"!=":          OPERATOR,
		"<=":          OPERATOR,
		">=":          OPERATOR,
		":=":          OPERATOR,
		"...":         OPERATOR,
	},
}
