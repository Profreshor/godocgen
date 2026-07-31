package evil

// a var that shares a name with the type below
var Parser = 1

type Parser struct{}

func (p *Parser) testnote()  {}
func (p *Parser) testnote2() {}
