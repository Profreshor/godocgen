package parser_test

import (
	"slices"
	"testing"

	"github.com/Profreshor/godocgen/internal/parser"
)

func TestAssembleOrphanNotes(t *testing.T) {
	orphans := []parser.Symbol{
		{Kind: parser.METHOD, Name: "M1", Owner: "Alpha"},
		{Kind: parser.METHOD, Name: "M2", Owner: "Beta"},
		{Kind: parser.METHOD, Name: "M3", Owner: "Gamma"},
		{Kind: parser.METHOD, Name: "M4", Owner: "Delta"},
	}
	want := []string{
		"orphaned method M1 references unknown type Alpha",
		"orphaned method M2 references unknown type Beta",
		"orphaned method M4 references unknown type Delta",
		"orphaned method M3 references unknown type Gamma",
	}
	_, got := parser.Assemble(orphans)

	if !slices.Equal(got, want) {
		t.Errorf("Orphan notes:\n got: %v\n want: %v", got, want)
	}
}
