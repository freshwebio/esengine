package parser

import (
	"fmt"
	"math"
	"strconv"
	"unicode"
)

// CommentType provides a type alias
// to distinguish types of comment tokens.
type CommentType int

const (
	_ CommentType = iota
	NonComment
	SingleLineComment
	MultiLineComment
)

const (
	MaxPunctuatorLength   = 4
	MaxReservedWordLength = 10
)

func IsLineTerminator(c rune, charMap map[string]map[rune]rune) bool {
	_, isLT := charMap["lineTerminators"][c]
	return isLT
}

// IsStartOfComment determines whether the given code point
// is the beginning of a comment.
func IsStartOfComment(c rune, next rune) (bool, CommentType) {
	if c == '/' && next == '/' {
		return true, SingleLineComment
	} else if c == '/' && next == '*' {
		return true, MultiLineComment
	}
	return false, NonComment
}

func IsStartOfIdentifier(c rune, pos int, buf []rune) (bool, int) {
	isEscapeSeq := false
	end := pos + 1
	if c == '\\' {
		isEscapeSeq, end = IsUnicodeEspaceSequence(pos+1, buf)
	}
	return IsIDStart(c) || c == '$' || c == '_' ||
		isEscapeSeq, end
}

func IsIdentifierPart(pos int, buf []rune) (bool, int) {
	isEscapeSeq := false
	end := pos + 1
	c := buf[pos]
	if c == '\\' {
		isEscapeSeq, end = IsUnicodeEspaceSequence(pos+1, buf)
	}
	return IsIDContinue(c) || c == '$' || c == '_' ||
		c == '\u200C' || c == '\u200D' || isEscapeSeq, end
}

func IsIDStart(c rune) bool {
	return (unicode.IsLetter(c) || unicode.In(c, unicode.Nl) ||
		unicode.In(c, unicode.Other_ID_Start)) &&
		!(unicode.In(c, unicode.Pattern_Syntax) || unicode.In(c, unicode.Pattern_White_Space))
}

func IsIDContinue(c rune) bool {
	return (IsIDStart(c) || unicode.In(c, unicode.Other_ID_Continue) ||
		unicode.In(c, unicode.Mn) || unicode.In(c, unicode.Mc) ||
		unicode.In(c, unicode.Nd) || unicode.In(c, unicode.Pc)) &&
		!(unicode.In(c, unicode.Pattern_Syntax) || unicode.In(c, unicode.Pattern_White_Space))
}

func IsUnicodeEspaceSequence(pos int, buf []rune) (bool, int) {
	// First code point should be a u.
	if buf[pos] != 'u' {
		return false, -1
	}
	// Make sure that there is a next code point.
	if pos+1 >= len(buf) {
		return false, -1
	}
	// If the next code point is not a { then the next 4 items
	// should be hex digits.
	if buf[pos+1] != '{' {
		// Ensure that there are 4 more code points
		// that follow at least.
		if pos+4 >= len(buf) {
			return false, -1
		}
		allAreHex := true
		i := 1
		for allAreHex && i < 5 {
			if !IsHexDigit(buf[pos+i]) {
				allAreHex = false
			} else {
				i++
			}
		}
		if !allAreHex {
			return false, -1
		}
		return true, pos + 5
	} else {
		reachedEnd := false
		allAreHex := true
		i := pos + 2
		for !reachedEnd && allAreHex && i < len(buf) {
			if buf[i] == '}' {
				reachedEnd = true
			} else {
				if !IsHexDigit(buf[i]) {
					allAreHex = false
				} else {
					i++
				}
			}
		}
		if !reachedEnd || !allAreHex {
			return false, -1
		}
		return true, i + 1
	}
}

func IsHexDigit(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') ||
		(c >= 'A' && c <= 'F')
}

// NextToken attempts to parse the next set of code points as a valid
// token in the ECMAScript lexical grammar.
func NextToken(pos int, buf []rune, charMap map[string]map[rune]rune,
	pMap map[string]rune, kwMap map[string]rune, frwMap map[string]rune) (*Token, int, error) {
	newPos := SkipNonToken(pos, buf, charMap)
	if newPos >= len(buf) {
		// In the case there are no tokens then simply return nil
		// for the token as well as the error as no tokens doesn't mean
		// there is an error.
		return nil, newPos, nil
	}
	c := buf[newPos]
	// Capture the next code point for cases we need to lookahead
	// to determine the type of token. Default to unicode 0 as when we do lookaheads
	// there is not a case where 0 is expected.
	next := '0'
	if newPos+1 < len(buf) {
		next = buf[newPos+1]
	}
	if tkn, endPos, err := ProcessDivPunctuator(newPos, buf); tkn != nil {
		return tkn, endPos, err
	} else if tkn, endPos, err := ProcessRightBracePunctuator(newPos, buf); tkn != nil {
		return tkn, endPos, err
	} else if IsLineTerminator(c, charMap) {
		return ProcessLineTerminator(newPos, buf)
	} else if isComment, commentType := IsStartOfComment(c, next); isComment {
		// Processing comment is effectively like skipping non-tokens
		// but we need to be more thorough as a comment that is multi-line and contains
		// a line terminator, it will need to be stored as a LineTerminator token for the sake of syntax parsing.
		return ProcessComment(newPos, buf, charMap, commentType)
	} else if tkn, endPos, err := ProcessPunctuator(newPos, buf, pMap); tkn != nil {
		// Since checking for a punctuator requires a bit more complexity than say,
		// a comment we are doing the checking and processing to generate a token in one.
		return tkn, endPos, err
	} else if isIDStart, endofStart := IsStartOfIdentifier(c, newPos, buf); isIDStart {
		return ProcessIdentifier(buf[newPos:endofStart], endofStart, buf)
	} else if tkn, endPos, err := ProcessReservedWord(newPos, buf, kwMap, frwMap); tkn != nil {
		return tkn, endPos, err
	} else if tkn, endPos, err := ProcessNumericLiteral(newPos, buf); tkn != nil {
		return tkn, endPos, err
	} else if tkn, endPos, err := ProcessRegExpLiteral(newPos, buf, charMap); tkn != nil {
		return tkn, endPos, err
	} else if tkn, endPos, err := ProcessStringLiteral(newPos, buf, charMap); tkn != nil {
		return tkn, endPos, err
	} else if tkn, endPos, err := ProcessTemplateLiteral(newPos, buf, charMap); tkn != nil {
		return tkn, endPos, err
	}
	return nil, newPos, fmt.Errorf("token error at %v for char %v", newPos, strconv.QuoteRune(c))
}

func SkipNonToken(pos int, buf []rune, charMap map[string]map[rune]rune) int {
	isWhiteSpace := true
	for isWhiteSpace && pos < len(buf) {
		_, isWhiteSpace = charMap["whitespace"][buf[pos]]
		if isWhiteSpace {
			pos++
		}
	}
	return pos
}

func WhiteSpaceChars() map[rune]rune {
	return map[rune]rune{
		'\u0009': 0,
		'\u000B': 0,
		'\u000C': 0,
		'\u00A0': 0,
		'\u0020': 0,
		// U+FEFF <ZWNBSP> is also recognised as a white space
		// character in ECMAScript.
		'\uFEFF': 0,
		// <USP> characters
		'\u1680': 0,
		'\u2000': 0,
		'\u2001': 0,
		'\u2002': 0,
		'\u2003': 0,
		'\u2004': 0,
		'\u2005': 0,
		'\u2006': 0,
		'\u2007': 0,
		'\u2008': 0,
		'\u2009': 0,
		'\u200A': 0,
		'\u202F': 0,
		'\u205F': 0,
		'\u3000': 0,
	}
}

func LineTerminators() map[rune]rune {
	return map[rune]rune{
		'\u000A': 0,
		'\u000D': 0,
		'\u2028': 0,
		'\u2029': 0,
	}
}

func Keywords() map[string]rune {
	return map[string]rune{
		"await": 0, "break": 0, "case": 0,
		"catch": 0, "class": 0, "const": 0, "continue": 0,
		"debugger": 0, "default": 0, "delete": 0, "do": 0,
		"else": 0, "export": 0, "extends": 0,
		"finally": 0, "for": 0, "function": 0,
		"if": 0, "import": 0, "in": 0, "instanceof": 0,
		"new": 0, "return": 0, "super": 0, "switch": 0,
		"this": 0, "throw": 0, "try": 0, "typeof": 0,
		"var": 0, "void": 0, "while": 0, "with": 0, "yield": 0,
	}
}

func FutureReservedWords() map[string]rune {
	return map[string]rune{
		"enum": 0,
	}
}

func Punctuators() map[string]rune {
	// Right brace punctuator and div punctuators are treated
	// differently and are not common tokens.
	return map[string]rune{
		"{": 0, "(": 0, ")": 0, "[": 0,
		"]": 0, ".": 0, "...": 0, ";": 0,
		",": 0, "<": 0, ">": 0, "<=": 0,
		">=": 0, "==": 0, "!=": 0, "===": 0,
		"!==": 0, "+": 0, "-": 0, "*": 0,
		"%": 0, "**": 0, "++": 0, "--": 0,
		"<<": 0, ">>": 0, ">>>": 0, "&": 0,
		"|": 0, "^": 0, "!": 0, "~": 0,
		"&&": 0, "||": 0, "?": 0, ":": 0,
		"=": 0, "+=": 0, "-=": 0, "*=": 0,
		"%=": 0, "**=": 0, "<<=": 0, ">>=": 0,
		">>>=": 0, "&=": 0, "|=": 0, "^=": 0,
		"=>": 0,
	}
}

// ProcessLineTerminator reads line terminator code points into the
// lexical analysis token table.
func ProcessLineTerminator(pos int, buf []rune) (*Token, int, error) {
	// Allow for <CR><LF> as a single source character.
	lineTerminator := string(buf[pos])
	endPos := pos + 1
	if pos+1 < len(buf) && (buf[pos] == '\u000D' && buf[pos+1] == '\u000A') {
		lineTerminator = string(buf[pos]) + string(buf[pos+1])
		endPos++
	}
	tkn := &Token{
		Name:  "LineTerminator",
		Value: lineTerminator,
		Pos:   pos,
	}
	return tkn, endPos, nil
}

// ProcessComment deals with creating a token for a comment
// that starts from position of the provided slice of code points.
// pos is including the two opening terminals.
func ProcessComment(pos int, buf []rune, charMap map[string]map[rune]rune, commentType CommentType) (*Token, int, error) {
	startPos := pos
	// Add two to bypass the opening terminal symbols.
	pos = pos + 2
	if commentType == SingleLineComment {
		isLineTerminator := false
		for !isLineTerminator && pos < len(buf) {
			_, isLineTerminator = charMap["lineTerminators"][buf[pos]]
			if !isLineTerminator {
				pos++
			}
		}
		return nil, pos, nil
	} else if commentType == MultiLineComment {
		comment := "/*"
		containsLineTerminator := false
		isEnd := false
		for !isEnd && pos < len(buf)-1 { // len(buf) - 1 as the last two code points make up the end.
			isEnd = buf[pos] == '*' && buf[pos+1] == '/'
			if !isEnd {
				if !containsLineTerminator && IsLineTerminator(buf[pos], charMap) {
					containsLineTerminator = true
				}
				comment += string(buf[pos])
				pos++
			}
		}
		if isEnd && !containsLineTerminator {
			// Add 2 to the position for the terminals */.
			return nil, pos + 2, nil
		} else if isEnd && containsLineTerminator {
			// Add to 2 to account for the comment block terminals
			// and that the end of the position is the start of the next token (or non-token).
			return &Token{
				Name:  "LineTerminator",
				Value: comment + "*/",
				Pos:   startPos,
			}, pos + 2, nil
		}
	}
	// We shouldn't ever get here, unless a comment is not terminated and we reach the end of the buffer.
	return nil, -1, fmt.Errorf("reached end of buffer and comment token was not closed")
}

func ProcessPunctuator(pos int, buf []rune, pMap map[string]rune) (*Token, int, error) {
	liFloat := math.Min(float64(pos+MaxPunctuatorLength), float64(len(buf)-1))
	lastIndex := int(liFloat)
	candidate := string(buf[pos:lastIndex])
	punctuator := candidate
	match := false
	i := len(candidate)
	for i > pos && !match {
		punctuator = candidate[pos:i]
		_, match = pMap[punctuator]
		if !match {
			i--
		}
	}
	if match {
		endPos := pos + len(punctuator)
		return &Token{
			Name:  "Punctuator",
			Value: punctuator,
			Pos:   pos,
		}, endPos, nil
	}
	return nil, -1, nil
}

func ProcessDivPunctuator(pos int, buf []rune) (*Token, int, error) {
	c := buf[pos]
	if c == '/' {
		if pos+1 < len(buf) {
			next := buf[pos+1]
			if next == '=' {
				return &Token{
					Name:  "DivPunctuator",
					Value: "/=",
					Pos:   pos,
				}, pos + 2, nil
			}
		}
		return &Token{
			Name:  "DivPunctuator",
			Value: "/",
			Pos:   pos,
		}, pos + 1, nil
	}
	return nil, -1, nil
}

func ProcessRightBracePunctuator(pos int, buf []rune) (*Token, int, error) {
	if buf[pos] == '}' {
		return &Token{
			Name:  "RightBracePunctuator",
			Value: "}",
			Pos:   pos,
		}, pos + 1, nil
	}
	return nil, -1, nil
}

func ProcessReservedWord(pos int, buf []rune, kwMap map[string]rune, frwMap map[string]rune) (*Token, int, error) {
	liFloat := math.Min(float64(pos+MaxReservedWordLength), float64(len(buf)-1))
	lastIndex := int(liFloat)
	candidate := string(buf[pos:lastIndex])
	reservedWord := candidate
	isKeyword := false
	isFutureReservedWord := false
	isBooleanLiteral := false
	isNullLiteral := false
	tokenType := ""
	i := len(candidate)
	for i > pos && !(isKeyword || isFutureReservedWord || isBooleanLiteral || isNullLiteral) {
		reservedWord = candidate[pos:i]
		_, isKeyword = kwMap[reservedWord]
		if !isKeyword {
			_, isFutureReservedWord = frwMap[reservedWord]
			if !isFutureReservedWord {
				isBooleanLiteral = reservedWord == "false" || reservedWord == "true"
				if !isBooleanLiteral {
					isNullLiteral = reservedWord == "null"
					if isNullLiteral {
						tokenType = "NullLiteral"
					} else {
						i-- // Reduce size of candidate as no match for current candidate.
					}
				} else {
					tokenType = "BooleanLiteral"
				}
			} else {
				tokenType = "FutureReservedWord"
			}
		} else {
			tokenType = "Keyword"
		}
	}
	if isKeyword || isFutureReservedWord || isBooleanLiteral || isNullLiteral {
		endPos := pos + len(reservedWord)
		return &Token{
			Name:  tokenType,
			Value: reservedWord,
			Pos:   pos,
		}, endPos, nil
	}
	return nil, -1, nil
}

func ProcessIdentifier(identifier []rune, pos int, buf []rune) (*Token, int, error) {
	// First ensure that what we have of an identifier so far is a valid
	// code point in the case it is a unicode escape sequence.
	if len(identifier) > 1 {
		codePoints := DecodeUnicodeEscSeq(identifier)
		isCodePointPart, _ := IsStartOfIdentifier(codePoints[0], 0, codePoints)
		if isCodePointPart {
			identifier = []rune{}
			identifier = append(identifier, codePoints...)
		} else {
			return nil, -1, fmt.Errorf("syntax error (early stage): " +
				"unicode escape sequence must represent a valid IdentifierStart")
		}
	}
	reachedEnd := false
	i := pos
	idEndPos := pos + 1
	for !reachedEnd && i < len(buf) {
		if isPart, endPos := IsIdentifierPart(pos, buf); isPart {
			if endPos > i+1 {
				// The only way an identifier part has a length greater
				// than one code point is where it is a unicode escape sequence.
				// In the light of that we'll replace the escape sequence with the
				// code point it represents.
				codePoints := DecodeUnicodeEscSeq(buf[pos:endPos])
				// Now our decoded code point must meet the same rules as any other
				// code point in an identifier part.
				isCodePointPart, _ := IsIdentifierPart(0, codePoints)
				if isCodePointPart {
					identifier = append(identifier, codePoints[0])
				} else {
					return nil, -1, fmt.Errorf("syntax error (early stage): " +
						"unicode escape sequence must represent a valid IdentifierPart")
				}
			} else {
				identifier = append(identifier, buf[i])
			}
			i++
			idEndPos += endPos
		} else {
			reachedEnd = true
		}
	}
	if reachedEnd {
		return &Token{
			Name:  "IdentifierName",
			Value: string(identifier),
			Pos:   pos,
		}, idEndPos, nil
	}
	return nil, -1, fmt.Errorf("identifier should terminate")
}

// ProcessNumericLiteral attempts to process the next set of
// code points as a numeric literal token.
func ProcessNumericLiteral(pos int, buf []rune) (*Token, int, error) {
	if tkn, endPos, err := ProcessDecimalLiteral(pos, buf); tkn != nil {
		return tkn, endPos, err
	} else if tkn, endPos, err := ProcessBinaryIntegerLiteral(pos, buf); tkn != nil {
		return tkn, endPos, err
	} else if tkn, endPos, err := ProcessOctalIntegerLiteral(pos, buf); tkn != nil {
		return tkn, endPos, err
	} else if tkn, endPos, err := ProcessHexIntegerLiteral(pos, buf); tkn != nil {
		return tkn, endPos, err
	}
	return nil, -1, nil
}

// ProcessDecimalLiteral attempts to process a decimal literal value.
func ProcessDecimalLiteral(pos int, buf []rune) (*Token, int, error) {
	reachedEnd := false
	i := pos
	intLiteral := ""
	decimalDigits := ""
	exponentPart := ""
	// The current production:
	// { DecimalIntegerLiteral = 1, DecimalDigits = 2, ExponentPart = 3 }
	currentProd := 1
	for !reachedEnd && i < len(buf) {
		if i == pos {
			// Allow for zero only when followed directly by
			// an exponent or decimal point.
			next := '0'
			if i+1 < len(buf) {
				next = buf[i+1]
			}
			if IsDecimalDigit(buf[i], true) || buf[i] == '0' && next == '.' {
				intLiteral += string(buf[i])
			} else if buf[i] == '.' {
				currentProd = 2
			} else {
				reachedEnd = true
			}
		} else if IsDecimalDigit(buf[i], false) {
			if currentProd == 1 {
				intLiteral += string(buf[i])
			} else if currentProd == 2 {
				decimalDigits += string(buf[i])
			} else if currentProd == 3 {
				exponentPart += string(buf[i])
			}
		} else if buf[i] == '.' {
			currentProd = 2
		} else if (buf[i] == 'e' || buf[i] == 'E') &&
			((intLiteral != "" && currentProd == 1) ||
				(decimalDigits != "" && currentProd == 2)) {
			currentProd = 3
		} else if (buf[i] == '-' || buf[i] == '+') && exponentPart == "" {
			exponentPart += string(buf[i])
		} else {
			reachedEnd = true
		}
		i++
	}
	if intLiteral == "" && decimalDigits == "" && exponentPart == "" {
		return nil, pos, nil
	}
	value := ""
	if intLiteral != "" {
		value += intLiteral
	}
	if decimalDigits != "" {
		value += "." + decimalDigits
	}
	if exponentPart != "" {
		value += "e" + exponentPart
	}
	tkn := &Token{
		Name:  "DecimalLiteral",
		Value: value,
		Pos:   pos,
	}
	// The end of the current token is always the last checked position
	// if we reached the end.
	if reachedEnd {
		i--
	}
	return tkn, i, nil
}

// IsDecimalDigit determines whether the provided rune is
// a decimal digit unicode code point or not.
// If nonZero is true, then excludes 0, otherwise it is included.
func IsDecimalDigit(c rune, nonZero bool) bool {
	if nonZero {
		return c >= '1' && c <= '9'
	}
	return c >= '0' && c <= '9'
}

// ProcessBinaryIntegerLiteral deals with extracting a binary integer
// literal from the provided input.
func ProcessBinaryIntegerLiteral(pos int, buf []rune) (*Token, int, error) {
	reachedEnd := false
	binaryValue := ""
	i := pos
	if i == pos {
		if i+1 >= len(buf) {
			return nil, pos, nil
		}
		if buf[i] != '0' || !(buf[i+1] == 'b' || buf[i+1] == 'B') {
			return nil, pos, nil
		}
	}
	i += 2
	// Increment i by 2 to capture the actual binary digits.
	for !reachedEnd && i < len(buf) {
		if buf[i] == '0' || buf[i] == '1' {
			binaryValue += string(buf[i])
		} else {
			reachedEnd = true
		}
		i++
	}
	// The end of the current token is always the last checked position
	// if we reached the end.
	if reachedEnd {
		i--
	}
	if binaryValue != "" {
		return &Token{
			Name:  "BinaryIntegerLiteral",
			Value: binaryValue,
			Pos:   pos,
		}, i, nil
	}
	return nil, pos, nil
}

// ProcessOctalIntegerLiteral deals with extracting an octal integer
// literal from the input rune slice.
func ProcessOctalIntegerLiteral(pos int, buf []rune) (*Token, int, error) {
	reachedEnd := false
	octalValue := ""
	i := pos
	if i == pos {
		if i+1 >= len(buf) {
			return nil, pos, nil
		}
		if buf[i] != '0' || !(buf[i+1] == 'o' || buf[i+1] == 'O') {
			return nil, pos, nil
		}
	}
	i += 2
	// Increment i by 2 to capture the actual binary digits.
	for !reachedEnd && i < len(buf) {
		if IsOctalDigit(buf[i]) {
			octalValue += string(buf[i])
		} else {
			reachedEnd = true
		}
		i++
	}
	// The end of the current token is always the last checked position
	// if we reached the end.
	if reachedEnd {
		i--
	}
	if octalValue != "" {
		return &Token{
			Name:  "OctalIntegerLiteral",
			Value: octalValue,
			Pos:   pos,
		}, i, nil
	}
	return nil, pos, nil
}

// IsOctalDigit determines whether the provided rune is
// a octal digit unicode code point or not.
func IsOctalDigit(c rune) bool {
	return c >= '0' && c <= '7'
}

// ProcessHexIntegerLiteral attempts to extract a hexadecimal
// integer literal value to be added to the lexical token table
// from the provided input.
func ProcessHexIntegerLiteral(pos int, buf []rune) (*Token, int, error) {
	reachedEnd := false
	hexValue := ""
	i := pos
	if i == pos {
		if i+1 >= len(buf) {
			return nil, pos, nil
		}
		if buf[i] != '0' || !(buf[i+1] == 'x' || buf[i+1] == 'X') {
			return nil, pos, nil
		}
	}
	i += 2
	// Increment i by 2 to capture the actual binary digits.
	for !reachedEnd && i < len(buf) {
		if IsHexDigit(buf[i]) {
			hexValue += string(buf[i])
		} else {
			reachedEnd = true
		}
		i++
	}
	// The end of the current token is always the last checked position
	// if we reached the end.
	if reachedEnd {
		i--
	}
	if hexValue != "" {
		return &Token{
			Name:  "HexIntegerLiteral",
			Value: hexValue,
			Pos:   pos,
		}, i, nil
	}
	return nil, pos, nil
}

// ProcessStringLiteral attempts to process code points from the
// position provided in the input buffer onwards.
func ProcessStringLiteral(pos int, buf []rune, charMap map[string]map[rune]rune) (*Token, int, error) {
	// Represents the production
	// { DoubleStringCharacters = 1, SingleStringCharacters = 2 }.
	stringProd := 0
	stringVal := ""
	if buf[pos] == '\'' {
		stringProd = 2
	} else if buf[pos] == '"' {
		stringProd = 1
	} else {
		return nil, pos, nil
	}
	i := pos
	i++
	reachEnd := false
	for !reachEnd && i < len(buf) {
		isValidChar := false
		newPos := pos
		if stringProd == 1 {
			newPos, isValidChar = IsDoubleStringCharacter(i, buf, charMap)
		} else if stringProd == 2 {
			newPos, isValidChar = IsSingleStringCharacter(i, buf, charMap)
		}
		if isValidChar {
			stringVal += string(buf[i:newPos])
			i = newPos
		} else if (stringProd == 1 && buf[i] == '"') ||
			(stringProd == 2 && buf[i] == '\'') {
			reachEnd = true
			i++
		} else {
			return nil, -1, fmt.Errorf("invalid code point in string literal")
		}
	}
	if !reachEnd {
		return nil, -1, fmt.Errorf("string literals must have a terminating quote")
	}
	return &Token{
		Name:  "StringLiteral",
		Value: stringVal,
		Pos:   pos,
	}, i, nil
}

// IsDoubleStringCharacter determines whether the given
// rune is a valid double quotes string character.
func IsDoubleStringCharacter(pos int, buf []rune, charMap map[string]map[rune]rune) (int, bool) {
	isDSChar := false
	isLineContinuation := false
	newPos, isStrEscapeSeq := IsStringEscapeSequence(pos, buf, charMap)
	if !isStrEscapeSeq {
		newPos, isLineContinuation = IsLineContinuation(pos, buf, charMap)
	}
	if isStrEscapeSeq {
		isDSChar = true
	} else if isLineContinuation {
		isDSChar = true
	} else if !(buf[pos] == '"' || buf[pos] == '\\' || IsLineTerminator(buf[pos], charMap)) {
		isDSChar = true
		newPos++
	}
	return newPos, isDSChar
}

// IsSingleStringCharacter determines whether the given
// rune is valid single quotes string character.
func IsSingleStringCharacter(pos int, buf []rune, charMap map[string]map[rune]rune) (int, bool) {
	isSSChar := false
	isLineContinuation := false
	newPos, isStrEscapeSeq := IsStringEscapeSequence(pos, buf, charMap)
	if !isStrEscapeSeq {
		newPos, isLineContinuation = IsLineContinuation(pos, buf, charMap)
	}
	if isStrEscapeSeq {
		isSSChar = true
	} else if isLineContinuation {
		isSSChar = true
	} else if !(buf[pos] == '\'' || buf[pos] == '\\' || IsLineTerminator(buf[pos], charMap)) {
		isSSChar = true
		newPos++
	}
	return newPos, isSSChar
}

// IsStringEscapeSequence determines whether the next set of characters
// from the given position is that of a valid string escape sequence.
func IsStringEscapeSequence(pos int, buf []rune, charMap map[string]map[rune]rune) (int, bool) {
	if buf[pos] == '\\' {
		newPos := pos + 1
		isStrEscapeSeq := false
		// In the case the next code point is 0 and the next after
		// that is a digit then that's our escape sequence.
		if pos+1 < len(buf) && buf[newPos] == '0' &&
			IsDecimalDigit(buf[newPos], false) {
			newPos++
			isStrEscapeSeq = true
		} else if isUnicodeEscapeSeq, endPos := IsUnicodeEspaceSequence(newPos, buf); isUnicodeEscapeSeq {
			newPos = endPos
			isStrEscapeSeq = true
		} else if isHexEscapeSeq, endPos := IsHexEscapeSequence(newPos, buf); isHexEscapeSeq {
			newPos = endPos
			isStrEscapeSeq = true
		} else if newPos < len(buf) &&
			(IsSingleEscapeCharacter(buf[newPos]) || IsNonEscapeCharacter(buf[newPos], charMap)) {
			newPos++
			isStrEscapeSeq = true
		}
		return newPos, isStrEscapeSeq
	}
	return pos, false
}

// IsHexEscapeSequence determines whether the next set of characters
// from the given position is that of a valid hexedecimal escape sequence.
func IsHexEscapeSequence(pos int, buf []rune) (bool, int) {
	if pos+2 < len(buf) && buf[pos] == 'x' &&
		IsHexDigit(buf[pos+1]) && IsHexDigit(buf[pos+2]) {
		return true, pos + 2
	}
	return false, pos
}

// IsLineContinuation determines whether the next set of characters
// from the given position is that of a valid line continuation.
func IsLineContinuation(pos int, buf []rune, charMap map[string]map[rune]rune) (int, bool) {
	if buf[pos] == '\\' {
		return IsLineTerminatorSequence(pos+1, buf, charMap)
	}
	return pos, false
}

// IsLineTerminatorSequence determines whether the next set of
// characters form a line terminator sequence.
func IsLineTerminatorSequence(pos int, buf []rune, charMap map[string]map[rune]rune) (int, bool) {
	endPos := pos + 1
	isLTSeq := false
	if pos+1 < len(buf) && (buf[pos] == '\u000D' && buf[pos+1] == '\u000A') {
		isLTSeq = true
		endPos++
	} else if IsLineTerminator(buf[pos], charMap) {
		isLTSeq = true
	}
	return endPos, isLTSeq
}

// IsSingleEscapeCharacter determines whether the provided code point
// is a single escape character.
func IsSingleEscapeCharacter(c rune) bool {
	return c == '\'' || c == '"' || c == '\\' || c == 'b' || c == 'f' ||
		c == 'n' || c == 'r' || c == 't' || c == 'v'
}

// IsNonEscapeCharacter determines whether the provided code point
// is a non-escape character.
func IsNonEscapeCharacter(c rune, charMap map[string]map[rune]rune) bool {
	return !IsEscapeCharacter(c) && !IsLineTerminator(c, charMap)
}

// IsEscapeCharacter determines whether the provided code point is an escape
// character as per the ECMAScript 8 specification.
func IsEscapeCharacter(c rune) bool {
	return !IsSingleEscapeCharacter(c) &&
		!IsDecimalDigit(c, false) && c != 'x' && c != 'u'
}

// ProcessRegExpLiteral attempts to process the next sequence of code points
// as a regular expression literal.
func ProcessRegExpLiteral(pos int, buf []rune, charMap map[string]map[rune]rune) (*Token, int, error) {
	if buf[pos] != '/' {
		return nil, pos, nil
	}
	isREBody, nextPos := IsRegExpBody(pos+1, buf, charMap)
	if !isREBody {
		return nil, pos, nil
	}
	if buf[nextPos] != '/' {
		return nil, pos, fmt.Errorf("Invalid regular expression literal missing closing /")
	}
	reBody := string(buf[(pos + 1):nextPos])
	prevPos := nextPos + 1
	_, nextPos = IsRegExpFlags(nextPos+1, buf)
	// Since regular expression flags can be empty if the next character
	// is not a valid flag then we finish the regexp literal before then.
	reFlags := string(buf[prevPos:nextPos])
	return &Token{
		Name:  "RegularExpressionLiteral",
		Value: "/" + reBody + "/" + reFlags,
		Pos:   pos,
	}, nextPos, nil
}

// IsRegExpBody determines whether the next sequence of
// code points make up a regular expression literal body.
func IsRegExpBody(pos int, buf []rune, charMap map[string]map[rune]rune) (bool, int) {
	// First ensure the first character is valid.
	firstIsValid, nextPos := IsFirstRegExpChar(pos, buf, charMap)
	if !firstIsValid {
		return false, pos
	}
	reachedEnd := false
	i := nextPos
	for !reachedEnd && i < len(buf) {
		var isValid bool
		isValid, nextPos = IsRegExpChar(i, buf, charMap)
		if isValid {
			i = nextPos
		} else {
			reachedEnd = true
		}
	}
	return reachedEnd, i
}

// IsRegExpFlags determines whether the next sequence of code
// points make up a regular expression literal flags section.
// This takes a naive approach in assuming the flags section
// has ended as soon as a non-identifier part code point is met.
func IsRegExpFlags(pos int, buf []rune) (bool, int) {
	if pos >= len(buf) {
		// We have reached the end of the buffer
		// so we assume empty flags.
		return true, pos
	}
	reachedEnd := false
	i := pos
	for !reachedEnd && i < len(buf) {
		isIdentifierPart, nextPos := IsIdentifierPart(i, buf)
		if !isIdentifierPart {
			reachedEnd = true
		} else {
			i = nextPos
		}
	}
	// If we have reached the end of the buffer
	// then we have reached the end of the regular expression flags.
	if i >= len(buf) {
		reachedEnd = true
	}
	return reachedEnd, i
}

// IsFirstRegExpChar determines whether the given code point
// is a valid regular expression first character.
func IsFirstRegExpChar(pos int, buf []rune, charMap map[string]map[rune]rune) (bool, int) {
	isRegExpBackSlashSeq, nextPos := IsRegExpBackSlashSeq(pos, buf, charMap)
	isRegExpClass := false
	if !isRegExpBackSlashSeq {
		isRegExpClass, nextPos = IsRegExpClass(pos, buf, charMap)
	}
	isRegExpNonTerminator := false
	if !isRegExpClass {
		isRegExpNonTerminator = IsRegExpNonTerminator(buf[pos], charMap) &&
			!(buf[pos] == '*' || buf[pos] == '\\' || buf[pos] == '/' || buf[pos] == '[')
		if isRegExpNonTerminator {
			nextPos = pos + 1
		}
	}
	return isRegExpNonTerminator ||
		isRegExpBackSlashSeq || isRegExpClass, nextPos
}

// IsRegExpChar determines whether the given code point
// is a valid regular expression character.
func IsRegExpChar(pos int, buf []rune, charMap map[string]map[rune]rune) (bool, int) {
	isRegExpBackSlashSeq, nextPos := IsRegExpBackSlashSeq(pos, buf, charMap)
	isRegExpClass := false
	if !isRegExpBackSlashSeq {
		isRegExpClass, nextPos = IsRegExpClass(pos, buf, charMap)
	}
	isRegExpNonTerminator := false
	if !isRegExpClass {
		isRegExpNonTerminator = IsRegExpNonTerminator(buf[pos], charMap) &&
			!(buf[pos] == '\\' || buf[pos] == '/' || buf[pos] == '[')
		if isRegExpNonTerminator {
			nextPos = pos + 1
		}
	}
	return isRegExpNonTerminator ||
		isRegExpBackSlashSeq || isRegExpClass, nextPos
}

// IsRegExpNonTerminator determines whether the provided code point
// is a valid regular expression non-terminator.
func IsRegExpNonTerminator(c rune, charMap map[string]map[rune]rune) bool {
	return !IsLineTerminator(c, charMap)
}

// IsRegExpBackSlashSeq determines whether the provided code point
// is the beginning of regular expression back slash sequence.
func IsRegExpBackSlashSeq(pos int, buf []rune, charMap map[string]map[rune]rune) (bool, int) {
	if buf[pos] == '\\' {
		if pos+1 < len(buf) && IsRegExpNonTerminator(buf[pos+1], charMap) {
			return true, pos + 2
		}
	}
	return false, pos
}

// IsRegExpClass determines whether the next set of code points makes up
// a valid regular expression class.
func IsRegExpClass(pos int, buf []rune, charMap map[string]map[rune]rune) (bool, int) {
	if buf[pos] == '[' {
		// An empty set of chars is valid.
		if pos+1 < len(buf) && buf[pos+1] == ']' {
			return true, pos + 2
		}
		isRegExpClassChars, nextPos := IsRegExpClassChars(pos+1, buf, charMap)
		if isRegExpClassChars && buf[nextPos] == ']' {
			return true, nextPos + 1
		}
	}
	return false, pos
}

// IsRegExpClassChars determines whether the next code point or set of
// code points represents a valid regular expression class characters sequence.
func IsRegExpClassChars(pos int, buf []rune, charMap map[string]map[rune]rune) (bool, int) {
	reachedEnd := false
	i := pos
	for !reachedEnd && i < len(buf) {
		isValid, nextPos := IsRegExpClassChar(i, buf, charMap)
		if isValid {
			i = nextPos
		} else {
			reachedEnd = true
		}
	}
	return reachedEnd, i
}

// IsRegExpClassChar determines whether the next sequence of code points
// make up a valid regular expression character.
func IsRegExpClassChar(pos int, buf []rune, charMap map[string]map[rune]rune) (bool, int) {
	isRegExpBackSlashSeq := false
	var nextPos int
	isNonTerminator := IsRegExpNonTerminator(buf[pos], charMap) &&
		!(buf[pos] == ']' || buf[pos] == '\\')
	if isNonTerminator {
		nextPos = pos + 1
	} else {
		isRegExpBackSlashSeq, nextPos = IsRegExpBackSlashSeq(pos, buf, charMap)
	}
	return isNonTerminator || isRegExpBackSlashSeq, nextPos
}

// ProcessTemplateLiteral attempts to parse the next sequence of code points
// from the provided position as a template literal.
func ProcessTemplateLiteral(pos int, buf []rune, charMap map[string]map[rune]rune) (*Token, int, error) {
	if buf[pos] == '`' {
		nextPos := ReadTemplateChars(pos+1, buf, charMap)
		if nextPos < len(buf) && buf[nextPos] == '`' {
			return &Token{
				Name:  "NoSubstitionTemplate",
				Value: string(buf[pos+1 : nextPos]),
				Pos:   pos,
			}, nextPos + 1, nil
		} else if nextPos+1 < len(buf) && buf[nextPos] == '$' &&
			buf[nextPos+1] == '{' {
			value := ""
			if nextPos-1 >= pos+1 {
				value = string(buf[pos+1 : nextPos-1])
			}
			return &Token{
				Name:  "TemplateHead",
				Value: value,
				Pos:   pos,
			}, nextPos + 2, nil
		}
	} else if buf[pos] == '}' {
		nextPos := ReadTemplateChars(pos+1, buf, charMap)
		if nextPos < len(buf) && buf[nextPos] == '`' {
			return &Token{
				Name:  "TemplateTail",
				Value: string(buf[pos+1 : nextPos]),
				Pos:   pos,
			}, nextPos + 1, nil
		} else if nextPos+1 < len(buf) && buf[nextPos] == '$' &&
			buf[nextPos+1] == '{' {
			value := ""
			if nextPos >= pos+1 {
				value = string(buf[pos+1 : nextPos])
			}
			return &Token{
				Name:  "TemplateMiddle",
				Value: value,
				Pos:   pos,
			}, nextPos + 1, nil
		}
	}
	return nil, pos, nil
}

// ReadTemplateChars parses the next sequence of code points
// from the specified position that makes up a set of valid TemplateCharacters.
func ReadTemplateChars(pos int, buf []rune, charMap map[string]map[rune]rune) int {
	reachedEnd := false
	i := pos
	for !reachedEnd && i < len(buf) {
		isTemplateChar, nextPos := IsTemplateChar(i, buf, charMap)
		if !isTemplateChar {
			reachedEnd = true
		} else {
			i = nextPos
		}
	}
	// A valid template chars can be empty so as soon as we hit a code point
	// which is not a valid template character then we can assume
	// we have reached the end of the sequence.
	return i
}

// IsTemplateChar determines whether the next code point or sequence of
// code points makes up a valid template character.
func IsTemplateChar(pos int, buf []rune, charMap map[string]map[rune]rune) (bool, int) {
	nextPos := pos
	isDollarNotSub := buf[pos] == '$' &&
		((pos+1 < len(buf) && buf[pos+1] != '{') || pos+1 >= len(buf))
	if isDollarNotSub {
		nextPos = pos + 1
	}
	isEscapeSeq := false
	isLineContinuation := false
	isLineTerminatorSequence := false
	isSourceCharWithExclusions := false
	if !isDollarNotSub {
		nextPos, isEscapeSeq = IsStringEscapeSequence(pos, buf, charMap)
		if !isEscapeSeq {
			nextPos, isLineContinuation = IsLineContinuation(pos, buf, charMap)
			if !isLineContinuation {
				nextPos, isLineTerminatorSequence = IsLineTerminatorSequence(pos, buf, charMap)
				if !isLineTerminatorSequence {
					isSourceCharWithExclusions = !(buf[pos] == '`' || buf[pos] == '\\' ||
						buf[pos] == '$' || IsLineTerminator(buf[pos], charMap))
					if isSourceCharWithExclusions {
						nextPos = pos + 1
					}
				}
			}
		}
	}
	return isDollarNotSub || isEscapeSeq || isLineContinuation ||
		isLineTerminatorSequence || isSourceCharWithExclusions, nextPos
}
