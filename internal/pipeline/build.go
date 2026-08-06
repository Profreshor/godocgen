package pipeline

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Profreshor/godocgen/internal/lexer"
	"github.com/Profreshor/godocgen/internal/parser"
	"github.com/Profreshor/godocgen/internal/walker"
)

func Build(fsys fs.FS, rootName string) ([]parser.PackageDoc, error) {
	project, err := walker.WalkFiles(fsys)
	if err != nil {
		return nil, err
	}
	piles := make(map[string][]parser.Symbol)
	for _, file := range project.Files {
		if file.LoadErr != nil {
			fmt.Fprintf(os.Stderr, "godocgen: skipping %s: %v\n", file.RelativePath, file.LoadErr)
			continue
		}
		if file.FileExt == ".go" && strings.HasSuffix(file.RelativePath, "_test.go") {
			continue
		}
		lex, err := lexer.CreateLexer(file.Content, file.FileExt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "godocgen: skipping %s: %v\n", file.RelativePath, err)
			continue
		}
		lex.Tokenize()
		p := parser.CreateParser(lex.Tokens, file.Content)
		p.Parse()
		syms := p.Symbols()
		for i := range syms {
			syms[i].File = file.RelativePath
			line, _ := lex.Position(syms[i].Span.Start.Byte)
			syms[i].Line = line
		}
		dir := filepath.Dir(file.RelativePath)
		if dir == "." {
			dir = rootName
		}
		piles[dir] = append(piles[dir], syms...)
	}

	dirs := make([]string, 0, len(piles))
	for dir := range piles {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	packages := make([]parser.PackageDoc, 0, len(dirs))
	for _, dir := range dirs {
		packages = append(packages, parser.AssemblePackage(dir, piles[dir]))
	}
	return packages, nil
}
