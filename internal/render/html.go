package render

import (
	"embed"
	"html/template"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Profreshor/godocgen/internal/parser"
)

//go:embed templates/*.html
var templateFS embed.FS

var pageTemplates = template.Must(
	template.New("pages").
		Funcs(template.FuncMap{
			"symbolID": symbolID,
			"cleanDoc": cleanDoc,
		}).
		ParseFS(templateFS, "templates/*.html"),
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

// --------- html template data ----------
type navItem struct {
	Slug  string
	Label string
}

type navGroup struct {
	Parent string
	Items  []navItem
}

type indexPage struct {
	Title string
	Nav   []navGroup
}

type packagePage struct {
	Title   string
	Path    string
	Doc     string
	Notes   []string
	Nav     []navGroup
	Groups  []kindGroup
	Imports []string
}

func RenderIndex(w io.Writer, title string, packages []parser.PackageDoc) error {
	page := indexPage{Title: title, Nav: buildNav(packages)}
	return pageTemplates.ExecuteTemplate(w, "index.html", page)
}

func RenderPackage(w io.Writer, pkg parser.PackageDoc, packages []parser.PackageDoc) error {
	page := packagePage{
		Title:   pkg.Name,
		Path:    pkg.Path,
		Doc:     cleanDoc(pkg.Doc),
		Notes:   pkg.Notes,
		Nav:     buildNav(packages),
		Groups:  groupAndSort(pkg.Symbols),
		Imports: collectImports(pkg.Symbols),
	}
	return pageTemplates.ExecuteTemplate(w, "package.html", page)
}

// --------- Helper Funcs ----------
type kindGroup struct {
	Kind    parser.SymbolKind
	Symbols []parser.Symbol
}

func Anchor(path string) string {
	return strings.ReplaceAll(path, "/", "-")
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

func buildNav(packages []parser.PackageDoc) []navGroup {
	buckets := make(map[string][]navItem)
	for _, pkg := range packages {
		parent := filepath.Dir(pkg.Path)
		if parent == "." {
			parent = "/"
		}
		buckets[parent] = append(buckets[parent], navItem{
			Slug:  Anchor(pkg.Path),
			Label: filepath.Base(pkg.Path),
		})
	}
	keys := make([]string, 0, len(buckets))
	for key := range buckets {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	groups := make([]navGroup, 0, len(keys))
	for _, key := range keys {
		groups = append(groups, navGroup{Parent: key, Items: buckets[key]})
	}
	return groups
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
