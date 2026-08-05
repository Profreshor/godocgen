package render

import (
	"fmt"
	"html"
	"io"
	"sort"
	"strings"

	"github.com/Profreshor/godocgen/internal/parser"
)

var kindOrder = []parser.SymbolKind{
	parser.CONSTANT,
	parser.VARIABLE,
	parser.FUNCTION,
	parser.TYPE,
	parser.INTERFACE,
	parser.STRUCT,
	parser.METHOD,
}

func writeIndex(pw *pageWriter, groups []kindGroup) {
	pw.writeLine(`<div class="index">`)
	for _, group := range groups {
		pw.printfLine(`<p><b>%s</b>`, html.EscapeString(group.Kind.String()))
		for i, s := range group.Symbols {
			if i > 0 {
				pw.write(", ")
			}
			writeIndexLink(pw, s)
			if len(s.Children) != 0 {
				pw.write(" (")
				for i := range s.Children {
					if i > 0 {
						pw.write(", ")
					}
					writeIndexLink(pw, s.Children[i])
				}
				pw.write(")")
			}
		}
		pw.writeLine(`</p>`)
	}
	pw.writeLine(`</div>`)
}

func writeIndexLink(pw *pageWriter, s parser.Symbol) {
	pw.printf(`<a href="#%s">%s</a>`, html.EscapeString(symbolID(s)), html.EscapeString(s.Name))
}

func writeHead(pw *pageWriter, title string) {
	pw.writeLine(`<!DOCTYPE html>`)
	pw.writeLine(`<html lang="en">`)
	pw.writeLine(`<head>`)
	pw.writeLine(`  <meta charset="UTF-8">`)
	pw.printfLine(`  <title>%s</title>`, html.EscapeString(title))
	pw.writeLine(`  <style>`)
	pw.writeLine(`    body { font-family: system-ui, sans-serif; line-height: 1.5; max-width: 900px; margin: 2rem auto; padding: 0 1rem; }`)
	pw.writeLine(`    .symbol { margin: 1.5rem 0; border-left: 3px solid #ddd; padding-left: 1rem; }`)
	pw.writeLine(`    .kind { color: #666; font-size: 0.85em; text-transform: uppercase; letter-spacing: 0.05em; }`)
	pw.writeLine(`    .name { font-weight: 600; }`)
	pw.writeLine(`    .detail { background: #f6f8fa; padding: 0.6rem 1rem; border-radius: 6px; overflow-x: auto; }`)
	pw.writeLine(`    .index { background: #f6f8fa; padding: 0.6rem 1rem; border-radius: 6px; overflow-x: auto; }`)
	pw.writeLine(`    .doc { margin-top: 0.75rem; white-space: pre-wrap; }`)
	pw.writeLine(`    .meta { color: #888; font-size: 0.8em; margin-top: 0.4rem; }`)
	pw.writeLine(`    ul.symbols { list-style: none; padding-left: 0; }`)
	pw.writeLine(`    ul.symbols ul.symbols { padding-left: 1.5rem; }`)
	pw.writeLine(`  </style>`)
	pw.writeLine(`</head>`)
	pw.writeLine(`<body>`)
}

func writeFoot(pw *pageWriter) {
	pw.writeLine(`</body>`)
	pw.writeLine(`</html>`)
}

func siteNav(pw *pageWriter, packages []parser.PackageDoc) {
	pw.writeLine(`<nav><ul>`)
	pw.writeLine(`<li><a href="/">⌂ index</a></li>`)
	for i := range packages {
		pw.printfLine(`  <li><a href="/pkg/%s">%s</a></li>`,
			Anchor(packages[i].Path),
			html.EscapeString(packages[i].Path))
	}
	pw.writeLine(`</ul></nav>`)
}

func RenderIndex(w io.Writer, title string, packages []parser.PackageDoc) error {
	pw := &pageWriter{destination: w}
	writeHead(pw, title)
	pw.printfLine(`<h1>%s</h1>`, html.EscapeString(title))
	siteNav(pw, packages)
	writeFoot(pw)
	return pw.firstError
}

func RenderPackage(w io.Writer, pkg parser.PackageDoc, packages []parser.PackageDoc) error {
	pw := &pageWriter{destination: w}
	writeHead(pw, pkg.Name)
	siteNav(pw, packages)
	pw.printfLine(`<section id="pkg-%s">`, Anchor(pkg.Path))
	pw.printfLine(`<h1>%s <span class="meta">%s</span></h1>`,
		html.EscapeString(pkg.Name),
		html.EscapeString(pkg.Path))
	if pkg.Doc != "" {
		pw.printfLine(`<p class="doc">%s</p>`, html.EscapeString(cleanDoc(pkg.Doc)))
	}
	for _, note := range pkg.Notes {
		pw.printfLine(`<div class="note">⚠ %s</div>`, html.EscapeString(note))
	}
	groups := groupAndSort(pkg.Symbols)
	writeIndex(pw, groups)
	pw.renderGroups(groups)
	imports := collectImports(pkg.Symbols)
	if len(imports) > 0 {
		pw.writeLine(`<h3>Imports</h3>`)
		pw.writeLine(`<ul class="imports">`)
		for _, name := range imports {
			pw.printfLine("<li>%s</li>", html.EscapeString(name))
		}
		pw.writeLine("</ul>")
	}
	pw.writeLine(`</section>`)
	writeFoot(pw)
	return pw.firstError
}

func Anchor(path string) string {
	return strings.ReplaceAll(path, "/", "-")
}

type pageWriter struct {
	destination io.Writer
	firstError  error
}

// write primitive
func (pw *pageWriter) write(text string) {
	if pw.firstError != nil {
		return
	}
	_, pw.firstError = io.WriteString(pw.destination, text)
}

// New line wrapper
func (pw *pageWriter) writeLine(text string) {
	pw.write(text + "\n")
}

// Formatting helper
func (pw *pageWriter) printf(format string, args ...any) {
	if pw.firstError != nil {
		return
	}
	_, pw.firstError = fmt.Fprintf(pw.destination, format, args...)
}

func (pw *pageWriter) printfLine(format string, args ...any) {
	pw.printf(format+"\n", args...)
}

// Escape helper
func (pw *pageWriter) writeEscaped(text string) {
	pw.write(html.EscapeString(text))
}

// --------- rendering ----------

func (pw *pageWriter) renderSymbols(symbols []parser.Symbol) {
	if len(symbols) == 0 {
		return
	}
	pw.writeLine(`<ul class="symbols">`)
	for i := range symbols {
		pw.renderSymbol(symbols[i])
	}
	pw.writeLine(`</ul>`)
}

func (pw *pageWriter) renderSymbol(s parser.Symbol) {
	pw.printfLine(`<li class="symbol" id="%s">`, html.EscapeString(symbolID(s)))

	pw.printfLine(`  <div><span class="kind">%s</span> <span class="name"><code>%s</code></span></div>`,
		html.EscapeString(s.Kind.String()),
		html.EscapeString(s.Name),
	)
	if s.Owner != "" {
		pw.printf(`  <div class="meta">owner: <code>%s</code></div>`, html.EscapeString(s.Owner))
	}
	if s.Detail != "" {
		pw.write(`  <pre class="detail"><code>`)
		pw.writeEscaped(s.Detail)
		pw.writeLine(`</code></pre>`)
	}
	if s.Doc != "" {
		pw.writeLine(`  <div class="doc">`)
		pw.writeEscaped(cleanDoc(s.Doc))
		pw.writeLine(`</div>`)
	}
	if s.File != "" && s.Line != 0 {
		pw.printfLine(`  <div class="meta">%s:%d</div>`, html.EscapeString(s.File), s.Line)
	}
	if len(s.Children) > 0 {
		pw.renderSymbols(s.Children)
	}
	pw.writeLine(`</li>`)
}

// --------- Group and Sort ----------
type kindGroup struct {
	Kind    parser.SymbolKind
	Symbols []parser.Symbol
}

func symbolID(s parser.Symbol) string {
	if s.Owner != "" {
		return s.Owner + "." + s.Name
	}
	return s.Kind.String() + "-" + s.Name
}

func collectImports(symbols []parser.Symbol) []string {
	unique := make(map[string]bool, len(symbols))
	for _, s := range symbols {
		if s.Kind == parser.PACKAGE {
			unique[s.Name] = true
		}
	}
	var pkgNames []string
	for n := range unique {
		pkgNames = append(pkgNames, n)
	}
	sort.Strings(pkgNames)
	return pkgNames
}

func (pw *pageWriter) renderGroups(groups []kindGroup) {
	for _, g := range groups {
		pw.printfLine(`<h3>%s</h3>`, html.EscapeString(g.Kind.String()))
		pw.renderSymbols(g.Symbols)
	}
}

func groupAndSort(symbols []parser.Symbol) []kindGroup {
	byKind := make(map[parser.SymbolKind][]parser.Symbol)
	for _, s := range symbols {
		if s.Kind == parser.MODULE || s.Kind == parser.PACKAGE {
			continue
		}
		byKind[s.Kind] = append(byKind[s.Kind], s)
	}
	for k := range byKind {
		sort.Slice(byKind[k], func(i, j int) bool {
			return byKind[k][i].Name < byKind[k][j].Name
		})
	}
	var groups []kindGroup
	for _, k := range kindOrder {
		if list, ok := byKind[k]; ok {
			groups = append(groups, kindGroup{Kind: k, Symbols: list})
		}
	}
	return groups
}

// --------- cleaning ----------
func cleanDoc(doc string) string {
	if doc == "" {
		return ""
	}
	lines := strings.Split(doc, "\n")
	for i, line := range lines {
		line = strings.TrimPrefix(line, "// ")
		line = strings.TrimPrefix(line, "//")
		lines[i] = line
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
