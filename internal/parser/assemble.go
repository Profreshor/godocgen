package parser

import (
	"fmt"
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
	for ownerName, methodList := range methodsByOwner {
		if len(methodList) == 0 {
			continue
		}
		topLevel = append(topLevel, methodList...)
		names := make([]string, len(methodList))
		for i, m := range methodList {
			names[i] = m.Name
		}
		var note string
		switch len(names) {
		case 1:
			note = fmt.Sprintf("orphaned method %s references unknown type %s", names[0], ownerName)
		default:
			note = fmt.Sprintf("orphaned methods %s reference unknown type %s", strings.Join(names, ", "), ownerName)
		}
		notes = append(notes, note)
	}
	return topLevel, notes
}
