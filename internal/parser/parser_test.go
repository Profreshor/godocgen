package parser_test

import (
	"strings"
	"testing"

	"github.com/Profreshor/godocgen/internal/lexer"
	"github.com/Profreshor/godocgen/internal/parser"
)

type symView struct {
	Kind  parser.SymbolKind
	Name  string
	Owner string
}

func parseSource(t *testing.T, src string) []parser.Symbol {
	t.Helper()
	lex, err := lexer.CreateLexer([]byte(src), ".go")
	if err != nil {
		t.Fatalf("CreateLexer: %v", err)
	}
	lex.Tokenize()
	p := parser.CreateParser(lex.Tokens, []byte(src))
	p.Parse()
	return p.Symbols()
}

func findSymbol(t *testing.T, syms []parser.Symbol, name string) parser.Symbol {
	t.Helper()
	for _, s := range syms {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("symbol %q not found", name)
	return parser.Symbol{}
}

func TestDocOnLastDeclaration(t *testing.T) {
	src := "package demo\n\n// LastVar is the final declaration.\nvar LastVar =1\n"
	got := findSymbol(t, parseSource(t, src), "LastVar")
	if !strings.Contains(got.Doc, "final declaration") {
		t.Errorf("doc lost on last declaration in file: got Doc = %q", got.Doc)
	}
}

func TestDocOnNonLastDeclaration(t *testing.T) {
	src := "package demo\n\n// LastVar is the final declaration.\nvar LastVar = 1\n\nfunc Trailing() {}\n"
	got := findSymbol(t, parseSource(t, src), "LastVar")
	if !strings.Contains(got.Doc, "final declaration") {
		t.Errorf("doc missing even with trailing declaration: got Doc = %q", got.Doc)
	}
}

func viewOf(syms []parser.Symbol) []symView {
	views := make([]symView, 0, len(syms))
	for _, s := range syms {
		views = append(views, symView{Kind: s.Kind, Name: s.Name, Owner: s.Owner})
	}
	return views
}

func TestParseTable(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []symView
	}{
		{
			name: "plain function",
			src:  "package demo\n\nfunc Hello() {}\n",
			want: []symView{
				{parser.MODULE, "demo", ""},
				{parser.FUNCTION, "Hello", ""},
			},
		},
		{
			name: "method, value receiver",
			src:  "func (b Buffer) Reset() {}",
			want: []symView{
				{parser.METHOD, "Reset", "Buffer"},
			},
		},
		{
			name: "method, pointer receiver",
			src:  "func (b *Buffer) Grow() {}",
			want: []symView{
				{parser.METHOD, "Grow", "Buffer"},
			},
		},
		{
			name: "grouped cont",
			src:  "const (\n\tA = 1\n\tB = 2\n)",
			want: []symView{
				{parser.CONSTANT, "A", ""},
				{parser.CONSTANT, "B", ""},
			},
		},
		{
			name: "multi-name spec",
			src:  "var x, y int",
			want: []symView{
				{parser.VARIABLE, "x", ""},
				{parser.VARIABLE, "y", ""},
			},
		},
		{
			name: "struct promotion",
			src:  "type Point struct {\n\tX int\n}",
			want: []symView{
				{parser.STRUCT, "Point", ""},
			},
		},
		{
			name: "generic type",
			src:  "type List[T any] struct{...}",
			want: []symView{
				{parser.STRUCT, "List", ""},
			},
		},
		{
			name: "generic method",
			src:  "func (l *List[T]) Add(v T) {}",
			want: []symView{
				{parser.METHOD, "Add", "List"},
			},
		},
		{
			name: "aliased import",
			src: `package demo
			import (
				j "encoding/json"
			)`,
			want: []symView{
				{parser.MODULE, "demo", ""},
				{parser.PACKAGE, "j", ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := viewOf(parseSource(t, tt.src))
			if len(got) != len(tt.want) {
				t.Fatalf("symbol count: got %d, want %d\n  got:  %v\n  want: %v",
					len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("symbol %d: got %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
