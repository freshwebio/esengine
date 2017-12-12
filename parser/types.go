package parser

// Token holds a token produced in the token table
// of the lexical analysis stage.
type Token struct {
	Name  string
	Value string
	Pos   int
}
