package lexer_test

import (
	"fmt"
	"testing"

	"github.com/Profreshor/godocgen/internal/lexer"
)

type tokenView struct {
	Kind lexer.Tokenkind
	Text string
}

var toy = lexer.Language{
	Name:        "Toy",
	LineComment: "#",
	Strings:     []lexer.StringSyntax{{Opener: '\'', Escapes: false}},
	Literals: map[string]lexer.Tokenkind{
		"type": lexer.KEYWORD,
		"var":  lexer.KEYWORD,
		"(":    lexer.PUNCT,
		"[":    lexer.PUNCT,
		"{":    lexer.PUNCT,
	},
}

func lexSource(t *testing.T, src string) []tokenView {
	t.Helper()
	lex, err := lexer.CreateLexer([]byte(src), ".go")
	if err != nil {
		t.Fatalf("CreateLexer: %v", err)
	}
	lex.Tokenize()
	views := make([]tokenView, 0, len(lex.Tokens))
	for _, tok := range lex.Tokens {
		views = append(views, tokenView{
			Kind: tok.Kind,
			Text: src[tok.Span.Start.Byte:tok.Span.End.Byte],
		})
	}
	return views
}

func lexToy(t *testing.T, src string) []tokenView {
	t.Helper()
	lex := lexer.NewLexer([]byte(src), toy)

	lex.Tokenize()
	views := make([]tokenView, 0, len(lex.Tokens))
	for _, tok := range lex.Tokens {
		views = append(views, tokenView{
			Kind: tok.Kind,
			Text: src[tok.Span.Start.Byte:tok.Span.End.Byte],
		})
	}
	return views
}

func TestTokenize(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []tokenView
	}{
		{
			name: "identifier",
			src:  "foo",
			want: []tokenView{{lexer.IDENT, "foo"}, {lexer.EOF, ""}},
		},
		{
			name: "keyword",
			src:  "func",
			want: []tokenView{{lexer.KEYWORD, "func"}, {lexer.EOF, ""}},
		},
		{
			name: "longest operator wins",
			src:  "a <<= b",
			want: []tokenView{
				{lexer.IDENT, "a"},
				{lexer.OPERATOR, "<<="},
				{lexer.IDENT, "b"},
				{lexer.EOF, ""},
			},
		},
		{
			name: "string with escaped quote",
			src:  `"a\"b"`,
			want: []tokenView{{lexer.STRING, `"a\"b"`}, {lexer.EOF, ""}},
		},
		{
			name: "Raw string",
			src:  "`abc`",
			want: []tokenView{{lexer.STRING, "`abc`"}, {lexer.EOF, ""}},
		},
		{
			name: "Line Comment",
			src:  "// hi\nx",
			want: []tokenView{
				{lexer.COMMENT, "// hi"},
				{lexer.IDENT, "x"},
				{lexer.EOF, ""}},
		},
		{
			name: "Block Comment",
			src:  "/* hi */",
			want: []tokenView{{lexer.COMMENT, "/* hi */"}, {lexer.EOF, ""}},
		},
		{
			name: "Number",
			src:  "4.2",
			want: []tokenView{{lexer.NUMBER, "4.2"}, {lexer.EOF, ""}},
		},
		{
			name: "longest operator wins - 2 char",
			src:  "a << b",
			want: []tokenView{
				{lexer.IDENT, "a"},
				{lexer.OPERATOR, "<<"},
				{lexer.IDENT, "b"},
				{lexer.EOF, ""},
			},
		},
		{
			name: "longest operator wins - 1 char",
			src:  "a < b",
			want: []tokenView{
				{lexer.IDENT, "a"},
				{lexer.OPERATOR, "<"},
				{lexer.IDENT, "b"},
				{lexer.EOF, ""},
			},
		},
		{
			name: "string ending in escape at EOF",
			src:  `var S = "abc\`,
			want: []tokenView{
				{lexer.KEYWORD, "var"},
				{lexer.IDENT, "S"},
				{lexer.OPERATOR, "="},
				{lexer.ILLEGAL, `"abc\`},
				{lexer.EOF, ""},
			},
		},
		{
			name: "unicode identifier",
			src:  "añadir",
			want: []tokenView{{lexer.IDENT, "añadir"}, {lexer.EOF, ""}},
		},
		{
			name: "CJK identifier",
			src:  "名前",
			want: []tokenView{{lexer.IDENT, "名前"}, {lexer.EOF, ""}},
		},
		{
			name: "unicode inside string already works",
			src:  `"héllo 🎉"`,
			want: []tokenView{{lexer.STRING, `"héllo 🎉"`}, {lexer.EOF, ""}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := lexSource(t, tc.src)
			if len(got) != len(tc.want) {
				t.Fatalf("token count: got %d, want %d\n  got:  %v\n  want: %v",
					len(got), len(tc.want), got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("token %d: got %v, want %v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestPosition(t *testing.T) {
	src := "ab\ncdé\nf"
	cases := []struct {
		offset   int
		wantLine int
		wantCol  int
	}{
		{0, 1, 1},
		{2, 1, 3},
		{3, 2, 1},
		{5, 2, 3},
		{7, 2, 5},
		{8, 3, 1},
	}
	for _, row := range cases {
		t.Run(fmt.Sprintf("offset_%d", row.offset), func(t *testing.T) {
			lex, err := lexer.CreateLexer([]byte(src), ".go")
			if err != nil {
				t.Fatalf("CreateLexer: %v", err)
			}
			line, col := lex.Position(row.offset)
			if line != row.wantLine || col != row.wantCol {
				t.Errorf("offset %d: got line %d, col %d; want line %d, col %d",
					row.offset, line, col, row.wantLine, row.wantCol)
			}
		})
	}
}

func TestLanguageLex(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []tokenView
	}{
		{
			name: "line comment test",
			src:  "# This is a python comment",
			want: []tokenView{
				{lexer.COMMENT, "# This is a python comment"},
				{lexer.EOF, ""},
			},
		},
		{
			name: "string test",
			src:  "'hi'",
			want: []tokenView{
				{lexer.STRING, "'hi'"},
				{lexer.EOF, ""},
			},
		},
		{
			name: "toy string ignores backslash",
			src:  "'a\\'",
			want: []tokenView{
				{lexer.STRING, "'a\\'"},
				{lexer.EOF, ""},
			},
		},
		{
			name: "keyword test",
			src:  "type",
			want: []tokenView{
				{lexer.KEYWORD, "type"},
				{lexer.EOF, ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lexToy(t, tt.src)
			if len(got) != len(tt.want) {
				t.Fatalf("token count: got %d, want %d\n  got:  %v\n  want: %v",
					len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("token %d: got %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
