package render

import (
	"reflect"
	"testing"

	"github.com/Profreshor/godocgen/internal/parser"
)

func TestBuildNavRootParent(t *testing.T) {
	packages := []parser.PackageDoc{{Path: "lexer"}, {Path: "cmd/api"}}
	got := buildNav(packages)

	want := []navGroup{
		{Parent: "/", Items: []navItem{{Slug: "lexer", Label: "lexer"}}},
		{Parent: "cmd", Items: []navItem{{Slug: "cmd-api", Label: "api"}}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot=%+v\nwant=%+v", got, want)
	}
}
