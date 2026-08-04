package walker_test

import (
	"testing"
	"testing/fstest"

	"github.com/Profreshor/godocgen/internal/walker"
)

func TestWalkFiles(t *testing.T) {
	// Arrange
	fsys := fstest.MapFS{
		"keep/main.go":        {Data: []byte("package keep")},
		"keep/notes.txt":      {Data: []byte("prose, not code")},
		"vendor/dep/dep.go":   {Data: []byte("package dep")},
		"node_modules/x/x.go": {Data: []byte("package x")},
	}
	// Act
	project, err := walker.WalkFiles(fsys)
	if err != nil {
		t.Fatalf("WalkFiles: %v", err)
	}
	// Assert
	if len(project.Files) != 1 {
		t.Fatalf("captured %d files, want 1: %+v", len(project.Files), project.Files)
	}
	got := project.Files[0]
	if got.RelativePath != "keep/main.go" {
		t.Errorf("RelativePath = %q, want %q", got.RelativePath, "keep/main.go")
	}
	if got.FileExt != ".go" {
		t.Errorf("FileExt = %q, want %q", got.FileExt, ".go")
	}
	if got.LoadErr != nil {
		t.Errorf("LoadErr = %v, want nil", got.LoadErr)
	}
	if string(got.Content) != "package keep" {
		t.Errorf("Content = %q, want %q", got.Content, "package keep")
	}
}
