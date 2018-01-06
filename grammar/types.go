package grammar

import (
	"errors"
)

var (
	// ErrInvalidRightHandSide provides the error for the case when a right-hand side
	// set of rules is not of the expected form.
	ErrInvalidRightHandSide = errors.New("invalid form for a set of right-hand side rules")
)

// RHSRuleSymbol provides the base definition for all right
// hand side rules that make up a production.
type RHSRuleSymbol interface {
	// Name provides the name of the given rule.
	// The name of a non-terminal symbol or the value of a terminal,
	// or the name of a terminal symbol which is in fact a non-terminal
	// symbol in the lexical grammar.
	Name() string
	// Params provides the set of parameters used for an RHSRuleSymbol
	// instantiation to determine rules.
	Params() interface{}
}

// NonTerminalRHSRuleSymbol provides the right hand side
// rule for a non-terminal symbol.
type NonTerminalRHSRuleSymbol struct {
	name   string
	params *NtRHSParams
}

func (r *NonTerminalRHSRuleSymbol) Name() string {
	return r.name
}

func (r *NonTerminalRHSRuleSymbol) Params() interface{} {
	return r.params
}

type NtRHSParams struct {
	Passthrough []string `yaml:"passthrough,omitempty"`
	Conditions  []string `yaml:"conditions,omitempty"`
	Optional    *bool    `yaml:"optional,omitempty"`
}

type TerminalRHSRuleSymbol struct {
	name   string
	params *TRHSParams
}

func (r *TerminalRHSRuleSymbol) Name() string {
	return r.name
}

func (r *TerminalRHSRuleSymbol) Params() interface{} {
	return r.params
}

type TRHSParams struct {
	Conditions []string `yaml:"conditions,omitempty"`
}

type ConditionalRHSRuleSymbol struct {
	params *CRHSParams
	Parts  []RHSRuleSymbol
}

func (r *ConditionalRHSRuleSymbol) Name() string {
	return ""
}

func (r *ConditionalRHSRuleSymbol) Params() interface{} {
	return r.params
}

// CRHSParams provides the parameters used for conditional right
// hand side rule.
type CRHSParams struct {
	Conditions []string
}

// ExcludeRHSRuleSymbol simply provides the data structure
// for exclusion of terminal symbols.
// Only the ECMAScript Syntactic grammar terminals are supported
// to be used by the exclude rule.
// To note, lexical grammar non-terminals are Syntactic grammar terminals.
type ExcludeRHSRuleSymbol struct {
	name string
}

func (r *ExcludeRHSRuleSymbol) Name() string {
	return r.name
}

func (r *ExcludeRHSRuleSymbol) Params() interface{} {
	return nil
}

type LookaheadRHSRuleSymbol struct {
	params *LaRHSParams
}

func (r *LookaheadRHSRuleSymbol) Name() string {
	return ""
}

func (r *LookaheadRHSRuleSymbol) Params() interface{} {
	return r.params
}

type LaRHSParams struct {
	Exclude [][]RHSRuleSymbol
}
