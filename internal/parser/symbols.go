package parser

import "github.com/Profreshor/godocgen/internal/lexer"

type Symbol struct {
	Kind     SymbolKind
	Name     string
	Span     lexer.Span
	Selector lexer.Span
	Doc      string
	Detail   string
	Children []Symbol
}

type SymbolKind uint8

const (
	FILE SymbolKind = iota
	MODULE
	// NAMESPACE
	PACKAGE
	// CLASS
	METHOD
	// PROPERTY
	// FIELD
	// CONSTRUCTOR
	// ENUM
	INTERFACE
	FUNCTION
	VARIABLE
	CONSTANT
	// STRING
	// NUMBER
	// BOOLEAN
	// ARRAY
	// OBJECT
	// KEY
	// NULL
	// ENUM_MEMBER
	STRUCT
	// EVENT
	// OPERATOR
	TYPE
)

func (s SymbolKind) String() string {
	switch s {
	case FILE:
		return "file"
	case MODULE:
		return "module"
	case PACKAGE:
		return "package"
	case METHOD:
		return "method"
	case FUNCTION:
		return "function"
	case VARIABLE:
		return "variable"
	case CONSTANT:
		return "constant"
	case TYPE:
		return "type"
	case STRUCT:
		return "struct"
	case INTERFACE:
		return "interface"
	default:
		return "unknown"
	}
}
