package render

import (
	"fmt"
	"html"
	"io"
	"sort"

	"github.com/Profreshor/godocgen/internal/parser"
)

var kindOrder = []parser.SymbolKind{
	parser.MODULE,
	parser.PACKAGE,
	parser.CONSTANT,
	parser.VARIABLE,
	parser.FUNCTION,
	parser.TYPE,
	parser.INTERFACE,
	parser.STRUCT,
	parser.METHOD,
}

func Render(w io.Writer, tree []parser.Symbol) error {
	pw := &pageWriter{destination: w}

	pw.writeLine(`<!DOCTYPE html>`)
	pw.writeLine(`<html lang="en">`)
	pw.writeLine(`<head>`)
	pw.writeLine(`  <meta charset="UTF-8">`)
	pw.writeLine(`  <title>godocgen</title>`)
	pw.writeLine(`  <style>`)
	pw.writeLine(`    body { font-family: system-ui, sans-serif; line-height: 1.5; max-width: 900px; margin: 2rem auto; padding: 0 1rem; }`)
	pw.writeLine(`    .symbol { margin: 1.5rem 0; border-left: 3px solid #ddd; padding-left: 1rem; }`)
	pw.writeLine(`    .kind { color: #666; font-size: 0.85em; text-transform: uppercase; letter-spacing: 0.05em; }`)
	pw.writeLine(`    .name { font-weight: 600; }`)
	pw.writeLine(`    .detail { background: #f6f8fa; padding: 0.6rem 1rem; border-radius: 6px; overflow-x: auto; }`)
	pw.writeLine(`    .doc { margin-top: 0.75rem; white-space: pre-wrap; }`)
	pw.writeLine(`    .meta { color: #888; font-size: 0.8em; margin-top: 0.4rem; }`)
	pw.writeLine(`    ul.symbols { list-style: none; padding-left: 0; }`)
	pw.writeLine(`    ul.symbols ul.symbols { padding-left: 1.5rem; }`)
	pw.writeLine(`  </style>`)
	pw.writeLine(`</head>`)
	pw.writeLine(`<body>`)

	groups := groupAndSort(tree)
	pw.renderGroups(groups)

	pw.writeLine(`</body>`)
	pw.writeLine(`</html>`)

	return pw.firstError
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

// Escape helper
func (pw *pageWriter) writeEscaped(text string) {
	pw.write(html.EscapeString(text))
}

// --------- rendering ----------

func (pw *pageWriter) renderSymbols(symbols []parser.Symbol) {
	if len(symbols) == 0 {
		return
	}
	pw.writeLine(`<ul class="symbols>`)
	for i := range symbols {
		pw.renderSymbol(&symbols[i])
	}
	pw.writeLine(`</ul>`)
}

func (pw *pageWriter) renderSymbol(s *parser.Symbol) {
	pw.writeLine(`<li class="symbol>`)

	pw.printf(`  <div><span class="kind">%s</span> <span class="name"><code>%s</code></span></div>`,
		html.EscapeString(s.Kind.String()),
		html.EscapeString(s.Name),
	)
	if s.Owner != "" {
		pw.printf(`  <div class="meta">owner: <code>%s</code></div>`, html.EscapeString(s.Owner))
	}
	if s.Detail != "" {
		pw.writeLine(`  <pre class="detail"><code>`)
		pw.writeEscaped(s.Detail)
		pw.writeLine(`</code></pre>`)
	}
	if s.Doc != "" {
		pw.writeLine(`  <div class="doc">`)
		pw.writeEscaped(s.Doc)
		pw.writeLine(`</div>`)
	}
	if s.File != "" {
		pw.printf(`  <div class="meta">%s</div>`, html.EscapeString(s.File))
	}
	if len(s.Children) > 0 {
		pw.renderSymbols(s.Children)
	}
	pw.writeLine(`</li>`)
}

type kindGroup struct {
	Kind    parser.SymbolKind
	Symbols []parser.Symbol
}

func (pw *pageWriter) renderGroups(groups []kindGroup) {
	for _, g := range groups {
		pw.printf(`<h2 id="%s">%s</h2.`,
			html.EscapeString(g.Kind.String()),
			html.EscapeString(g.Kind.String()))
		pw.renderSymbols(g.Symbols)
	}
}

func groupAndSort(symbols []parser.Symbol) []kindGroup {
	byKind := make(map[parser.SymbolKind][]parser.Symbol)
	for _, s := range symbols {
		byKind[s.Kind] = append(byKind[s.Kind], s)
	}
	for k := range byKind {
		sort.Slice(byKind[k], func(i, j int) bool {
			return byKind[k][i].Name < byKind[k][j].Name
		})
	}
	var groups []kindGroup
	seen := make(map[parser.SymbolKind]bool)
	for _, k := range kindOrder {
		if list, ok := byKind[k]; ok {
			groups = append(groups, kindGroup{Kind: k, Symbols: list})
			seen[k] = true
		}
	}
	var extra []parser.SymbolKind
	for k := range byKind {
		if !seen[k] {
			extra = append(extra, k)
		}
	}
	sort.Slice(extra, func(i, j int) bool {
		return extra[i].String() < extra[j].String()
	})
	for _, k := range extra {
		groups = append(groups, kindGroup{Kind: k, Symbols: byKind[k]})
	}
	return groups
}
