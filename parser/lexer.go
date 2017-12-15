package parser

// Lexer provides the base definition
// for a service which deals with tokenising
// an input slice of code points.
type Lexer interface {
	Tokenise(input []rune) ([]*Token, error)
}

// NewLexer creates a new instance of the default
// lexer service.
func NewLexer() Lexer {
	charMap := map[string]map[rune]rune{
		"whitespace":      WhiteSpaceChars(),
		"lineTerminators": LineTerminators(),
	}
	return &lexerImpl{
		charMap,
		Punctuators(),
		Keywords(),
		FutureReservedWords(),
		[]rune{},
		[]*Token{},
	}
}

type lexerImpl struct {
	charMap       map[string]map[rune]rune
	pMap          map[string]rune
	kwMap         map[string]rune
	frwMap        map[string]rune
	currentInput  []rune
	currentTokens []*Token
}

// Tokenise deals with generating a list of tokens for the given input
// data.
func (l *lexerImpl) Tokenise(input []rune) ([]*Token, error) {
	l.currentInput = input
	var err error
	i := 0
	for err == nil && i < len(input) {
		var tkn *Token
		var nextPos int
		tkn, nextPos, err = NextToken(i, input, l.charMap, l.pMap, l.kwMap, l.frwMap)
		if err == nil {
			l.currentTokens = append(l.currentTokens, tkn)
			i = nextPos
		}
	}
	return l.currentTokens, err
}
