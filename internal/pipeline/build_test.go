package pipeline_test

import (
	"testing"
	"testing/fstest"

	"github.com/Profreshor/godocgen/internal/pipeline"
)

func TestBuildNamesRootPile(t *testing.T) {
	fsys := fstest.MapFS{
		"lexer.go": &fstest.MapFile{Data: []byte("package lexer\n\nfunc Hello() {}\n")},
	}
	packages, err := pipeline.Build(fsys, "lexer")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(packages) != 1 {
		t.Fatal("Build: len packages not equal to 1")
	}
	if packages[0].Path != "lexer" {
		t.Errorf(`Got: %q Want: "lexer"`, packages[0].Path)
	}
}
