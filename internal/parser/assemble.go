package parser

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func Assemble(symbols []Symbol) ([]Symbol, []string) {
	// Phase 0: Empty Containers to manage state
	methodsByOwner := make(map[string][]Symbol)
	topLevel := make([]Symbol, 0)
	seenModule := false
	// Phase 1: Partition
	for i := range symbols {
		switch symbols[i].Kind {
		case METHOD:
			methodsByOwner[symbols[i].Owner] = append(methodsByOwner[symbols[i].Owner], symbols[i])
		case MODULE:
			if !seenModule {
				topLevel = append(topLevel, symbols[i])
				seenModule = true
			}
		default:
			topLevel = append(topLevel, symbols[i])
		}
	}
	// Phase 2: Attach
	for i := range topLevel {
		switch topLevel[i].Kind {
		case TYPE, STRUCT, INTERFACE:
			topLevel[i].Children = append(topLevel[i].Children, methodsByOwner[topLevel[i].Name]...)
			delete(methodsByOwner, topLevel[i].Name)
		}
	}
	// Phase 3: Conclusion
	var notes []string
	ownerKeys := make([]string, 0, len(methodsByOwner))
	for key := range methodsByOwner {
		ownerKeys = append(ownerKeys, key)
	}
	sort.Strings(ownerKeys)

	for _, owner := range ownerKeys {
		if len(methodsByOwner[owner]) == 0 {
			continue
		}
		topLevel = append(topLevel, methodsByOwner[owner]...)
		names := make([]string, len(methodsByOwner[owner]))
		for i, m := range methodsByOwner[owner] {
			names[i] = m.Name
		}
		var note string
		switch len(names) {
		case 1:
			note = fmt.Sprintf("orphaned method %s references unknown type %s", names[0], owner)
		default:
			note = fmt.Sprintf("orphaned methods %s reference unknown type %s", strings.Join(names, ", "), owner)
		}
		notes = append(notes, note)
	}
	return topLevel, notes
}

type PackageDoc struct {
	Name    string
	Path    string
	Doc     string
	Symbols []Symbol
	Notes   []string
}

func AssemblePackage(path string, symbols []Symbol) PackageDoc {
	tree, notes := Assemble(symbols)
	name := filepath.Base(path)
	doc := ""
	for _, s := range symbols {
		if s.Kind == MODULE {
			name = s.Name
			if doc == "" && s.Doc != "" {
				doc = s.Doc
			}
		}
	}
	return PackageDoc{
		Name:    name,
		Path:    path,
		Doc:     doc,
		Symbols: tree,
		Notes:   notes,
	}
}
