package parser_test

import (
	"strings"
	"testing"

	"github.com/Profreshor/godocgen/internal/lexer"
	"github.com/Profreshor/godocgen/internal/parser"
)

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
