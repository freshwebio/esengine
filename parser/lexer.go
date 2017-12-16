package parser

// Lexer provides the base definition
// for a service which deals with tokenising
// an input slice of code points.
type Lexer interface {
	Tokenise(input []rune, goal LexicalGoalSymbol) ([]*Token, error)
	Reset()
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
func (l *lexerImpl) Tokenise(input []rune, goal LexicalGoalSymbol) ([]*Token, error) {
	l.currentInput = input
	var err error
	i := 0
	for err == nil && i < len(input) {
		var tkn *Token
		var nextPos int
		tkn, nextPos, err = NextToken(i, input, l.charMap, l.pMap, l.kwMap, l.frwMap, goal)
		if err == nil {
			if tkn != nil {
				l.currentTokens = append(l.currentTokens, tkn)
			}
			i = nextPos
		}
	}
	return l.currentTokens, err
}

// Reset deals with resetting the lexer and clearing the token
// table and current input data.
func (l *lexerImpl) Reset() {
	l.currentInput = []rune{}
	l.currentTokens = []*Token{}
}
