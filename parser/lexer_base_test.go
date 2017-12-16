package parser

import (
	"fmt"
	"testing"
)

type tokeniseTestData struct {
	input         []rune
	shouldSucceed bool
	expected      []*Token
	goal          LexicalGoalSymbol
}

func tokeniseTest(t *testing.T, lexer Lexer, testData []*tokeniseTestData) {
	for _, testItem := range testData {
		lexer.Reset()
		result, err := lexer.Tokenise(testItem.input, testItem.goal)
		if testItem.shouldSucceed && err != nil {
			t.Error(err)
		} else if !testItem.shouldSucceed && err == nil {
			t.Error("Tokenisation of provided input should have caused an error but succeeded")
		}
		for i, res := range result {
			if i >= len(testItem.expected) {
				t.Error("Exceeded the expected amount of tokens")
			}
			expected := testItem.expected[i]
			if res.Name != expected.Name || res.Value != expected.Value || res.Pos != expected.Pos {
				preprocessTokensForOutput(expected, res)
				t.Errorf("Did not get the expected token %+v but got the token %+v", expected, res)
			}
		}
	}
}

// Allows us to display a Unicode or escape sequence representation of invisible characters.
func preprocessTokensForOutput(tokens ...*Token) {
	for _, tkn := range tokens {
		tkn.Value = fmt.Sprintf("%+q", tkn.Value)
	}
}
