package grammar

import (
	"testing"
)

func TestLeftFactor(t *testing.T) {
	input, expected := loadGrammarFixture("lf1")
	LeftFactor(input)
	assertSame(t, *input, *expected)
	input, expected = loadGrammarFixture("lf2")
	LeftFactor(input)
	assertSame(t, *input, *expected)
	input, expected = loadGrammarFixture("lf3")
	LeftFactor(input)
	assertSame(t, *input, *expected)
	input, expected = loadGrammarFixture("lf4")
	LeftFactor(input)
	assertSame(t, *input, *expected)
	input, expected = loadGrammarFixture("lf5")
	LeftFactor(input)
	assertSame(t, *input, *expected)
}
