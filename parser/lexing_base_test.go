package parser

import "testing"

type testData struct {
	shouldSucceed bool
	buf           []rune
	endPos        int
	expected      string
	extra         map[string]interface{}
}

func processTest(t *testing.T, data []testData, tokenType string, process func(int, []rune) (*Token, int, error)) {
	for i := 0; i < len(data); i++ {
		tkn, endPos, err := process(0, data[i].buf)
		if data[i].shouldSucceed && err != nil {
			t.Error(err)
		}
		if data[i].shouldSucceed {
			if tkn == nil {
				t.Errorf("Expected valid %v token but got nil", tokenType)
			}
			if endPos != data[i].endPos {
				t.Errorf(
					"Expected end of token to be position %v but got %v",
					data[i].endPos,
					endPos,
				)
			}
			if tkn.Value != data[i].expected {
				t.Errorf(
					"Expected value to be %v but got %v",
					data[i].expected,
					tkn.Value,
				)
			}
			if tkn.Name != tokenType {
				t.Errorf("Expected token to be %v but got %v", tokenType, tkn.Name)
			}
		} else {
			if tkn != nil {
				t.Errorf("Expected a nil token but got %v", tkn)
			}
		}
	}
}

func processTestWithCharMap(
	t *testing.T, data []testData, tokenType string,
	charMap map[string]rune,
	process func(int, []rune, map[string]rune) (*Token, int, error),
) {
	for i := 0; i < len(data); i++ {
		tkn, endPos, err := process(0, data[i].buf, charMap)
		if data[i].shouldSucceed && err != nil {
			t.Error(err)
		}
		if data[i].shouldSucceed {
			if tkn == nil {
				t.Errorf("Expected valid %v token but got nil", tokenType)
			}
			if endPos != data[i].endPos {
				t.Errorf(
					"Expected end of token to be position %v but got %v",
					data[i].endPos,
					endPos,
				)
			}
			if tkn.Value != data[i].expected {
				t.Errorf(
					"Expected value to be %v but got %v",
					data[i].expected,
					tkn.Value,
				)
			}
			if tkn.Name != tokenType {
				t.Errorf("Expected token to be %v but got %v", tokenType, tkn.Name)
			}
		} else {
			if tkn != nil {
				t.Errorf("Expected a nil token but got %v", tkn)
			}
		}
	}
}

func processTestWithCharMaps(
	t *testing.T, data []testData, tokenType string,
	charMap map[string]map[rune]rune,
	process func(int, []rune, map[string]map[rune]rune) (*Token, int, error),
) {
	for i := 0; i < len(data); i++ {
		tkn, endPos, err := process(0, data[i].buf, charMap)
		if data[i].shouldSucceed && err != nil {
			t.Error(err)
		}
		if data[i].shouldSucceed {
			if tkn == nil {
				t.Errorf("Expected valid %v token but got nil", tokenType)
			}
			if endPos != data[i].endPos {
				t.Errorf(
					"Expected end of token to be position %v but got %v",
					data[i].endPos,
					endPos,
				)
			}
			if tkn.Value != data[i].expected {
				t.Errorf(
					"Expected value to be %v but got %v",
					data[i].expected,
					tkn.Value,
				)
			}
			if tkn.Name != tokenType {
				t.Errorf("Expected token to be %v but got %v", tokenType, tkn.Name)
			}
		} else {
			if tkn != nil {
				t.Errorf("Expected a nil token or an error but got %v", tkn)
			}
		}
	}
}

func processTestForComments(
	t *testing.T, data []testData,
	charMap map[string]map[rune]rune,
	process func(int, []rune, map[string]map[rune]rune, CommentType) (*Token, int, error),
) {
	for i := 0; i < len(data); i++ {
		tkn, endPos, err := process(
			0, data[i].buf, charMap, data[i].extra["commentType"].(CommentType),
		)
		if data[i].shouldSucceed && err != nil {
			t.Error(err)
		}
		if data[i].shouldSucceed {
			if tkn != nil && data[i].extra["commentType"].(CommentType) != MultiLineComment {
				t.Errorf("Expected a nil token but got a %v token", tkn.Name)
			} else if tkn != nil && tkn.Name != "LineTerminator" &&
				data[i].extra["commentType"].(CommentType) == MultiLineComment {
				t.Errorf("Expected LineTerminator token but got %v token", tkn.Name)
			}
			if endPos != data[i].endPos {
				t.Errorf(
					"Expected end of token to be position %v but got %v",
					data[i].endPos,
					endPos,
				)
			}
			if tkn != nil && tkn.Value != data[i].expected {
				t.Errorf(
					"Expected value to be %v but got %v",
					data[i].expected,
					tkn.Value,
				)
			}
			if tkn != nil && tkn.Name != "LineTerminator" {
				t.Errorf("Expected token to be LineTerminator but got %v", tkn.Name)
			}
		} else {
			if tkn != nil {
				t.Errorf("Expected a nil token but got %v", tkn)
			}
		}
	}
}
