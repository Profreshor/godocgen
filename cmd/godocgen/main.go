package main

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Profreshor/godocgen/internal/pipeline"
	"github.com/Profreshor/godocgen/internal/render"
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
	packages, err := pipeline.Build(os.DirFS(absPath), filepath.Base(absPath))
	if err != nil {
		return err
	}

	pages := make(map[string][]byte)
	var index bytes.Buffer
	if err := render.RenderIndex(&index, filepath.Base(absPath), packages); err != nil {
		return fmt.Errorf("render index: %w", err)
	}
	for _, pkg := range packages {
		var buf bytes.Buffer
		if err := render.RenderPackage(&buf, pkg, packages); err != nil {
			return fmt.Errorf("render %s: %w", pkg.Path, err)
		}
		pages[render.Anchor(pkg.Path)] = buf.Bytes()
	}

	fmt.Println(logo)
	fmt.Printf("godocgen: %d packages documented\n", len(packages))

	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(index.Bytes())
	})
	mux.HandleFunc("/pkg/{slug}", func(w http.ResponseWriter, r *http.Request) {
		body, ok := pages[r.PathValue("slug")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(body)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	fmt.Printf("godocgen: serving docs at http://%s\n", listener.Addr())
	return http.Serve(listener, mux)
}
