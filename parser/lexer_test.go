package parser

import (
	"testing"
)

func TestLexerTokenise(t *testing.T) {

	inputData := []*tokeniseTestData{
		{[]rune("//Some comment\nclass MyAwesomeClass\n { /* Constructor */\nconstructor() {} }"), true, []*Token{
			{Name: "LineTerminator", Value: "\n", Pos: 14},
			{Name: "Keyword", Value: "class", Pos: 15},
			{Name: "IdentifierName", Value: "MyAwesomeClass", Pos: 21},
			{Name: "LineTerminator", Value: "\n", Pos: 35},
			{Name: "Punctuator", Value: "{", Pos: 37},
			{Name: "LineTerminator", Value: "\n", Pos: 56},
			{Name: "IdentifierName", Value: "constructor", Pos: 57},
			{Name: "Punctuator", Value: "(", Pos: 68},
			{Name: "Punctuator", Value: ")", Pos: 69},
			{Name: "Punctuator", Value: "{", Pos: 71},
			{Name: "RightBracePunctuator", Value: "}", Pos: 72},
			{Name: "RightBracePunctuator", Value: "}", Pos: 74},
		}, InputElementDiv},
		{[]rune("/ab*.+\\+?/g \"My string literal \\n value\"}"), true, []*Token{
			{Name: "RegularExpressionLiteral", Value: "/ab*.+\\+?/g", Pos: 0},
			{Name: "StringLiteral", Value: "My string literal \\n value", Pos: 12},
			{Name: "RightBracePunctuator", Value: "}", Pos: 40},
		}, InputElementRegExp},
		{[]rune("/** Some comment text \n */\n/ab*.+\\+?/g"), false, []*Token{
			{Name: "LineTerminator", Value: "/** Some comment text \n */", Pos: 0},
			{Name: "LineTerminator", Value: "\n", Pos: 26},
			{Name: "DivPunctuator", Value: "/", Pos: 27},
			{Name: "IdentifierName", Value: "ab", Pos: 28},
			{Name: "Punctuator", Value: "*", Pos: 30},
			{Name: "Punctuator", Value: ".", Pos: 31},
			{Name: "Punctuator", Value: "+", Pos: 32},
		}, InputElementDiv},
		{[]rune("let myVar = 23 / 4;"), true, []*Token{
			{Name: "IdentifierName", Value: "let", Pos: 0},
			{Name: "IdentifierName", Value: "myVar", Pos: 4},
			{Name: "Punctuator", Value: "=", Pos: 10},
			{Name: "DecimalLiteral", Value: "23", Pos: 12},
			{Name: "DivPunctuator", Value: "/", Pos: 15},
			{Name: "DecimalLiteral", Value: "4", Pos: 17},
			{Name: "Punctuator", Value: ";", Pos: 18},
		}, InputElementDiv},
		{[]rune("let newVar /= 3;"), true, []*Token{
			{Name: "IdentifierName", Value: "let", Pos: 0},
			{Name: "IdentifierName", Value: "newVar", Pos: 4},
			{Name: "DivPunctuator", Value: "/=", Pos: 11},
			{Name: "DecimalLiteral", Value: "3", Pos: 14},
			{Name: "Punctuator", Value: ";", Pos: 15},
		}, InputElementDiv},
		{[]rune("let aVar = /^awQ+[A-Za-z]$/i;"), true, []*Token{
			{Name: "IdentifierName", Value: "let", Pos: 0},
			{Name: "IdentifierName", Value: "aVar", Pos: 4},
			{Name: "Punctuator", Value: "=", Pos: 9},
			{Name: "RegularExpressionLiteral", Value: "/^awQ+[A-Za-z]$/i", Pos: 11},
			{Name: "Punctuator", Value: ";", Pos: 28},
		}, InputElementRegExpOrTemplateTail},
		{[]rune("} Some template tail text`"), true, []*Token{
			{Name: "TemplateTail", Value: " Some template tail text", Pos: 0},
		}, InputElementRegExpOrTemplateTail},
		{[]rune("}Some more template tail text`"), true, []*Token{
			{Name: "TemplateTail", Value: "Some more template tail text", Pos: 0},
		}, InputElementTemplateTail},
		{[]rune("/ab*/g"), false, []*Token{
			{Name: "DivPunctuator", Value: "/", Pos: 0},
			{Name: "IdentifierName", Value: "ab", Pos: 1},
			{Name: "Punctuator", Value: "*", Pos: 3},
			{Name: "DivPunctuator", Value: "/", Pos: 4},
		}, InputElementTemplateTail},
	}
	lexer := NewLexer()
	tokeniseTest(t, lexer, inputData)
}
