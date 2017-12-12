package parser

import (
	"testing"
)

func TestProcessLineTerminator(t *testing.T) {
	var lineTerminatorData = []testData{
		{true, []rune{'\u000D', '\u000A'}, 2, "\u000D\u000A", nil},
		{true, append([]rune{'\u000D'}, []rune("Next line")...), 1, "\u000D", nil},
	}
	processTest(t, lineTerminatorData, "LineTerminator", ProcessLineTerminator)
}

func TestSkipNonToken(t *testing.T) {
	buf := append([]rune{'\u0009', '\u000C'}, []rune("Next line")...)
	charMap := map[string]map[rune]rune{
		"whitespace": WhiteSpaceChars(),
	}
	nextPos := SkipNonToken(0, buf, charMap)
	if nextPos != 2 {
		t.Errorf("Expected next position for a start of a token to be 2 but got %v", nextPos)
	}
	buf = []rune("Our one and only piece of content")
	nextPos = SkipNonToken(0, buf, charMap)
	if nextPos != 0 {
		t.Errorf("Expected next position for a start of a token to be 0 but got %v", nextPos)
	}
}

func TestProcessComment(t *testing.T) {
	var commentData = []testData{
		{true, []rune("//Comment1"), 10, "Comment1", map[string]interface{}{
			"commentType": SingleLineComment,
		}},
		{true, []rune("// My single line comment 2\nfunction someFunction() {}"), 27, "", map[string]interface{}{
			"commentType": SingleLineComment,
		}},
		{true, []rune("//"), 2, "", map[string]interface{}{
			"commentType": SingleLineComment,
		}},
		{true, []rune("/**\n* Comment Line 1\n* Comment Line 2*/"), 39,
			"/**\n* Comment Line 1\n* Comment Line 2*/", map[string]interface{}{
				"commentType": MultiLineComment,
			}},
		{true, []rune("/* This is our block comment */\nfunction myAwesomeFunction() {}"),
			31, "", map[string]interface{}{
				"commentType": MultiLineComment,
			}},
	}
	charMap := map[string]map[rune]rune{
		"lineTerminators": LineTerminators(),
	}
	processTestForComments(t, commentData, charMap, ProcessComment)
}

func TestProcessPunctuator(t *testing.T) {
	var punctuatorData = []testData{
		{true, []rune(">>>= fantastic 4"), 4, ">>>=", nil},
		{true, []rune("=== thorough 3"), 3, "===", nil},
		{true, []rune("<< terrible 2"), 2, "<<", nil},
		{true, []rune("% the one"), 1, "%", nil},
		{false, []rune("The one without a punctuator"), 0, "", nil},
	}
	pMap := Punctuators()
	processTestWithCharMap(t, punctuatorData, "Punctuator", pMap, ProcessPunctuator)
}

func TestProcessReservedWord(t *testing.T) {
}

func TestIsStartOfIdentifier(t *testing.T) {
}

func TestProcessIdentifier(t *testing.T) {
}

func TestProcessDecimalLiteral(t *testing.T) {
	var decimalData = []testData{
		{true, []rune("54.34E-23 nextValue"), 9, "54.34e-23", nil},
		{true, []rune(".57 "), 3, ".57", nil},
		{false, []rune(" "), 0, "", nil},
		{true, []rune("35"), 2, "35", nil},
		{true, []rune("35e+32"), 6, "35e+32", nil},
		{true, []rune(".75e-6021"), 9, ".75e-6021", nil},
		{true, []rune("0.49"), 4, "0.49", nil},
	}
	processTest(t, decimalData, "DecimalLiteral", ProcessDecimalLiteral)
}

func TestProcessBinaryIntegerLiteral(t *testing.T) {
	var binaryIntData = []testData{
		{true, []rune("0b011010001"), 11, "011010001", nil},
		{true, []rune("0B0"), 3, "0", nil},
		{false, []rune("0b"), 0, "", nil},
		{false, []rune("0B"), 0, "", nil},
		{true, []rune("0b1023"), 4, "10", nil},
		{false, []rune("0"), 0, "", nil},
		{false, []rune("ab"), 0, "", nil},
	}
	processTest(t, binaryIntData, "BinaryIntegerLiteral", ProcessBinaryIntegerLiteral)
}

func TestProcessOctalIntegerLiteral(t *testing.T) {
	var octalIntData = []testData{
		{true, []rune("0o014569"), 7, "01456", nil},
		{true, []rune("0o0"), 3, "0", nil},
		{false, []rune("0o"), 0, "", nil},
		{false, []rune("0O"), 0, "", nil},
		{true, []rune("0o1309"), 5, "130", nil},
		{false, []rune("0"), 0, "", nil},
		{false, []rune("ao"), 0, "", nil},
	}
	processTest(t, octalIntData, "OctalIntegerLiteral", ProcessOctalIntegerLiteral)
}

func TestProcessHexIntegerLiteral(t *testing.T) {
	var hexIntData = []testData{
		{true, []rune("0x014569FE2*"), 11, "014569FE2", nil},
		{true, []rune("0x0"), 3, "0", nil},
		{false, []rune("0x"), 0, "", nil},
		{false, []rune("0X"), 0, "", nil},
		{true, []rune("0X1309ATWQ"), 7, "1309A", nil},
		{false, []rune("0"), 0, "", nil},
		{false, []rune("ax"), 0, "", nil},
	}
	processTest(t, hexIntData, "HexIntegerLiteral", ProcessHexIntegerLiteral)
}

func TestProcessStringLiteral(t *testing.T) {
	var stringData = []testData{
		{true, []rune(`"\x99\u00AC2580"`), 16, `\x99\u00AC2580`, nil},
		{true, []rune(`'test single quotes'`), 20, `test single quotes`, nil},
		{false, []rune(`"invalid string`), 0, "", nil},
		{true, []rune("\"Test string\\\u000D\""), 15, "Test string\\\u000D", nil},
		{true, []rune("'Test string\\\u000D\u000A' some after text"), 16, "Test string\\\u000D\u000A", nil},
		{false, []rune(`'Invalid single quotes string`), 0, "", nil},
		{false, []rune(`var notAString = 2;`), 0, "", nil},
		{false, []rune("\"hey\u000D\""), 0, "", nil},
		{true, []rune("\"\\02 string\\'\""), 14, "\\02 string\\'", nil},
		{true, []rune("'hey\\0'"), 7, "hey\\0", nil},
	}
	charMap := map[string]map[rune]rune{
		"lineTerminators": LineTerminators(),
	}
	processTestWithCharMaps(t, stringData, "StringLiteral", charMap, ProcessStringLiteral)
}
