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
	NAMESPACE
	PACKAGE
	CLASS
	METHOD
	PROPERTY
	FIELD
	CONSTRUCTOR
	ENUM
	INTERFACE
	FUNCTION
	VARIABLE
	CONSTANT
	STRING
	NUMBER
	BOOLEAN
	ARRAY
	OBJECT
	KEY
	NULL
	ENUM_MEMBER
	STRUCT
	EVENT
	OPERATOR
	TYPE_PARAMETER
)
