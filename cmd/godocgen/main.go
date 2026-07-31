package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Profreshor/godocgen/internal/lexer"
	"github.com/Profreshor/godocgen/internal/parser"
	"github.com/Profreshor/godocgen/internal/render"
	"github.com/Profreshor/godocgen/internal/walker"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("godocgen: Please enter <project-directory>")
		os.Exit(1)
	}
	inputPath := os.Args[1]
	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		os.Exit(1)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("godocgen: %s: No such directory\n", inputPath)
			os.Exit(1)
		}
		fmt.Printf("godocgen: %s\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Println("godocgen: Please enter a directory path")
		os.Exit(1)
	}

	project, err := walker.WalkFiles(absPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// piles are groups of symbols per package
	piles := make(map[string][]parser.Symbol)
	for _, file := range project.Files {
		if file.LoadErr != nil {
			fmt.Printf("Skipping %s: due to load error: %v\n", file.RelativePath, file.LoadErr)
			continue
		}
		lex, err := lexer.CreateLexer(file.Content, file.FileExt)
		if err != nil {
			fmt.Printf("godocgen: %v\n", err)
			os.Exit(1)
		}
		lex.Tokenize()
		p := parser.CreateParser(lex.Tokens, file.Content)
		p.Parse()
		syms := p.Symbols()
		for i := range syms {
			syms[i].File = file.RelativePath
		}
		dir := filepath.Dir(file.RelativePath)
		piles[dir] = append(piles[dir], syms...)
	}
	dirs := make([]string, 0, len(piles))
	for dir := range piles {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	var packages []parser.PackageDoc
	for _, dir := range dirs {
		packages = append(packages, parser.AssemblePackage(dir, piles[dir]))
	}
	if err := os.MkdirAll("output", 0o755); err != nil {
		fmt.Printf("godocgen: %v\n", err)
		os.Exit(1)
	}
	f, err := os.Create("output/out.html")
	if err != nil {
		fmt.Printf("godocgen: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := render.Render(f, filepath.Base(absPath), packages); err != nil {
		fmt.Printf("godocgen: %v\n", err)
		os.Exit(1)
	}
	for _, pkg := range packages {
		fmt.Printf("== %s ==\n", pkg.Path)
		for _, n := range pkg.Notes {
			fmt.Println("note:", n)
		}
		for _, symbol := range pkg.Symbols {
			printSymbol(symbol, 0)
		}
	}
}

func printSymbol(s parser.Symbol, depth int) {
	indent := strings.Repeat("  ", depth)
	line := s.Detail
	if strings.Contains(line, "\n") {
		parts := strings.Split(line, "\n")
		line = parts[0] + "..."
	}
	if depth == 0 {
		fmt.Printf("%s%-9s %s  %s  [%s]\n", indent, s.Kind, s.Name, line, s.File)
	} else {
		fmt.Printf("%s%-9s %s  %s\n", indent, s.Kind, s.Name, line)
	}
	for _, child := range s.Children {
		printSymbol(child, depth+1)
	}
}
