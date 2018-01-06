package parser

//go:generate esegrammar build -grammar grammar.yml -output grammar.go -package parser
import (
	"bytes"
	"errors"
	"unicode/utf8"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
)

var (
	// ErrInvalidUnicodeSourceText provides the error when source text
	// contains characters which are not valid unicode code points.
	ErrInvalidUnicodeSourceText = errors.New("Non-unicode source text was present in the provided source text")
)

// Parser provides the base definition
// for a service that deals with parsing input
// into an AST.
type Parser interface {
	ParseScript([]byte, *RealmRecord, interface{}) *ScriptRecord
	ParseModule([]byte, *RealmRecord, interface{}) (ModuleRecord, error)
}

// NewParser creates a new instance of the default
// implementation of the parser.
func NewParser(lexer Lexer) Parser {
	return &parserImpl{
		lexer, false, InputElementDiv,
		ParseStack{}, map[Symbol]map[Symbol]int{},
	}
}

// Provides the default implementation
// of the parser.
type parserImpl struct {
	lexer        Lexer
	inStrictMode bool
	lexicalGoal  LexicalGoalSymbol
	parseStack   ParseStack
	parseTable   map[Symbol]map[Symbol]int
}

// Parse deals with parsing the given set
// of code points in to a parse tree.
func (p *parserImpl) ParseScript(sourceText []byte, realm *RealmRecord, hostDefined interface{}) *ScriptRecord {
	return nil
}

// ParseModule deals with attempting to parse the given input text
// as an ECMAScript module. only UTF-8, UTF-16 and UTF-32 source text inputs are supported,
// if the source file is encoded in any other format, then convert to UTF-8 before
// attempting to parse the module.
// UTF-8 is preferred as is the native unicode format of Go.
func (p *parserImpl) ParseModule(sourceText []byte, realm *RealmRecord, hostDefined interface{}) (ModuleRecord, error) {
	// In the case our source text is empty (Since ModuleBody is optional)
	// we are finished.
	if len(sourceText) == 0 {
		return &SourceTextModuleRecord{
			AbstractModuleRecord:  &AbstractModuleRecord{},
			ParseTree:             nil,
			RequestedModules:      []string{},
			ImportEntries:         []*ImportEntry{},
			LocalExportEntries:    []*ExportEntry{},
			IndirectExportEntries: []*ExportEntry{},
			StarExportEntries:     []*ExportEntry{},
		}, nil
	}
	isValidUnicode, decoded, err := p.validateSourceText(sourceText)
	if !isValidUnicode {
		return nil, ErrInvalidUnicodeSourceText
	} else if err != nil {
		return nil, err
	}
	input := bytes.Runes(decoded)
	errors := []error{}
	tree := (*ParseNode)(nil)
	p.parseModule(input, tree, errors)
	return nil, nil
}

func (p *parserImpl) parseModule(input []rune, tree *ParseNode, errors []error) {
}

// Deals with validating the given source code contains nothing but
// valid unicode source characters.
func (p *parserImpl) validateSourceText(sourceText []byte) (bool, []byte, error) {
	isValid := utf8.Valid(sourceText)
	decoded := sourceText
	if !isValid {
		// Get all unicode encodings apart from UTF-8.
		unicodeEncodings := append(unicode.All[1:], utf32.All...)
		i := 0
		for !isValid && i < len(unicodeEncodings) {
			decoder := unicodeEncodings[i].NewDecoder()
			var err error
			decoded, err = decoder.Bytes(sourceText)
			if err != nil {
				return false, nil, err
			}
			isValid = utf8.Valid(decoded)
			if !isValid {
				i++
			}
		}
	}
	return isValid, decoded, nil
}
