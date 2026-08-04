package parser_test

import (
	"slices"
	"testing"

	"github.com/Profreshor/godocgen/internal/parser"
)

func TestAssembleDeterministicNotes(t *testing.T) {
	orphans := []parser.Symbol{
		{Kind: parser.METHOD, Name: "M1", Owner: "Alpha"},
		{Kind: parser.METHOD, Name: "M2", Owner: "Beta"},
		{Kind: parser.METHOD, Name: "M3", Owner: "Gamma"},
		{Kind: parser.METHOD, Name: "M4", Owner: "Delta"},
	}
	_, notes1 := parser.Assemble(orphans)
	_, notes2 := parser.Assemble(orphans)

	if !slices.Equal(notes1, notes2) {
		t.Errorf("Assemble notes not deterministic:\n run 1: %v\n run 2: %v", notes1, notes2)
	}
}
