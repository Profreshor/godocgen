package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/Profreshor/godocgen/internal/lexer"
	"github.com/Profreshor/godocgen/internal/parser"
	"github.com/Profreshor/godocgen/internal/render"
	"github.com/Profreshor/godocgen/internal/walker"
)

const logo string = `
        ▌               
▞▀▌▞▀▖▞▀▌▞▀▖▞▀▖▞▀▌▞▀▖▛▀▖
▚▄▌▌ ▌▌ ▌▌ ▌▌ ▖▚▄▌▛▀ ▌ ▌
▗▄▘▝▀ ▝▀▘▝▀ ▝▀ ▗▄▘▝▀▘▘ ▘`

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "godocgen:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: godocgen <project-directory>")
	}
	absPath, err := filepath.Abs(args[1])
	if err != nil {
		return err
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", args[1])
	}
	packages, err := buildDocs(absPath)
	if err != nil {
		return err
	}
	var page bytes.Buffer
	if err := render.Render(&page, filepath.Base(absPath), packages); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	fmt.Println(logo)
	fmt.Printf("godocgen: %d packages documented\n", len(packages))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(page.Bytes())
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	fmt.Printf("godocgen: serving docs at http://%s\n", listener.Addr())
	return http.Serve(listener, mux)
}

func buildDocs(root string) ([]parser.PackageDoc, error) {
	project, err := walker.WalkFiles(os.DirFS(root))
	if err != nil {
		return nil, err
	}
	piles := make(map[string][]parser.Symbol)
	for _, file := range project.Files {
		if file.LoadErr != nil {
			fmt.Fprintf(os.Stderr, "godocgen: skipping %s: %v\n", file.RelativePath, file.LoadErr)
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
		}
		dir := filepath.Dir(file.RelativePath)
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
