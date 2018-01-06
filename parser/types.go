package parser

// Token holds a token produced in the token table
// of the lexical analysis stage.
type Token struct {
	Name  string
	Value string
	Pos   int
}

type Symbol int

// ParseNode represents a symbol in the
// parse tree.
type ParseNode struct {
	// The symbol of the current parse node.
	// For non-terminals, the goal symbol and for terminals
	// the terminal symbol value.
	Symbol Symbol
	// Terminal determines whether or not the current parse node
	// is a terminal symbol.
	Terminal bool
	// Children represents from left to right,
	// the child nodes of our current root node.
	Children []*ParseNode
}

// ParseStack provides a stack data structure
// used in parsing ECMAScript.
//
// This is non-threadsafe but that shouldn't be a problem
// as will only be used for the sequrntial process of parsing the
// ECMAScript (derivative) grammar.
type ParseStack []Symbol

// Push adds a token to the top of the stack.
func (s ParseStack) Push(v Symbol) ParseStack {
	return append(s, v)
}

// Pop takes a token from the top of the stack.
func (s ParseStack) Pop() (ParseStack, Symbol) {
	l := len(s)
	if l == 0 {
		return s, 0
	}
	return s[:l-1], s[l-1]
}

type LexicalEnvironment struct {
}

// RealmRecord provides the realm a script has been created in
// for script evaluation.
type RealmRecord struct {
	GlobalObject map[string]interface{}
}

// ScriptRecord provides the record which
// encapsulates information about the script being evaluated.
type ScriptRecord struct {
	ParseTree *ParseNode
	Errors    []error
	Realm     *RealmRecord
}

type ResolvedBindingRecord struct {
	Module      ModuleRecord
	BindingName string
}

type ResolveExportEntry struct {
	Module     ModuleRecord
	ExportName string
}

type ExportEntry struct {
	ExportName    string
	ModuleRequest string
	ImportName    string
	LocalName     string
}

type ImportEntry struct {
	ModuleRequest string
	ImportName    string
	LocalName     string
}

type ModuleRecord interface {
	GetExportedNames([]ModuleRecord) []string
	ResolveExport(string, []*ResolveExportEntry) (*ResolvedBindingRecord, string)
	ModuleDeclarationInstantiation()
	ModuleEvaluation()
}

type AbstractModuleRecord struct {
	ModuleRecord
	Realm       *RealmRecord
	Environment *LexicalEnvironment
	Namespace   map[string]interface{}
	Evaluated   bool
	HostDefined interface{}
}

type SourceTextModuleRecord struct {
	*AbstractModuleRecord
	ParseTree             *ParseNode
	RequestedModules      []string
	ImportEntries         []*ImportEntry
	LocalExportEntries    []*ExportEntry
	IndirectExportEntries []*ExportEntry
	StarExportEntries     []*ExportEntry
}

// GetExportedNames retrieves the exported names of the module
// given the provided export star set.
func (r *SourceTextModuleRecord) GetExportedNames(exportStarSet []ModuleRecord) []string {
	found := false
	i := 0
	for !found && i < len(exportStarSet) {
		if r == exportStarSet[i] {
			found = true
		} else {
			i++
		}
	}
	if found {
		return []string{}
	}
	exportStarSet = append(exportStarSet, r)
	return nil
}

func (r *SourceTextModuleRecord) ResolveExport(exportName string, resolveSet []*ResolveExportEntry) (*ResolvedBindingRecord, string) {
	return nil, ""
}

func (r *SourceTextModuleRecord) ModuleDeclarationInstantiation() {
}

func (r *SourceTextModuleRecord) ModuleEvaluation() {}
