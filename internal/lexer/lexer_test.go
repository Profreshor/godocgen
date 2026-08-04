package lexer_test

import (
	"testing"

	"github.com/Profreshor/godocgen/internal/lexer"
)

type tokenView struct {
	Kind lexer.Tokenkind
	Text string
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
