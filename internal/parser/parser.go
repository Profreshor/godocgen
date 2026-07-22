package parser

import "github.com/Profreshor/godocgen/internal/lexer"

type Parser struct {
	tokens  []lexer.Token
	source  []byte
	pos     int
	symbols []Symbol
}
