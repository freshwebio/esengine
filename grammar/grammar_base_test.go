package grammar

import (
	"io/ioutil"
	"reflect"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

// Loads a grammar fixture by name,
// returns input grammar followed by expected result grammar
func loadGrammarFixture(name string) (*Grammar, *Grammar) {
	inputBytes, err := ioutil.ReadFile("test/fixtures/" + name + "/input.yml")
	if err != nil {
		panic(err)
	}
	inputGrammar := &Grammar{}
	err = yaml.Unmarshal(inputBytes, inputGrammar)
	if err != nil {
		panic(err)
	}
	expectedBytes, err := ioutil.ReadFile("test/fixtures/" + name + "/expected.yml")
	if err != nil {
		panic(err)
	}
	expectedGrammar := &Grammar{}
	err = yaml.Unmarshal(expectedBytes, expectedGrammar)
	return inputGrammar, expectedGrammar
}

// Determines if the two given grammars are equivalent.
func isSameGrammar(g1 Grammar, g2 Grammar) bool {
	if len(g1.Productions) != len(g2.Productions) {
		return false
	}
	same := true
	i := 0
	for same && i < len(g1.Productions) {
		if g1.Productions[i].Name != g2.Productions[i].Name ||
			len(g1.Productions[i].Params) != len(g2.Productions[i].Params) ||
			len(g1.Productions[i].RHS) != len(g2.Productions[i].RHS) {
			same = false
		} else {
			j := 0
			for same && j < len(g1.Productions[i].Params) {
				if g1.Productions[i].Params[j] != g2.Productions[i].Params[j] {
					same = false
				} else {
					j++
				}
			}
			if same {
				j := 0
				for same && j < len(g1.Productions[i].RHS) {
					if len(g1.Productions[i].RHS[j]) != len(g2.Productions[i].RHS[j]) {
						same = false
					}
					k := 0
					for same && k < len(g1.Productions[i].RHS[j]) {
						same = isSameSymbol(g1.Productions[i].RHS[j][k], g2.Productions[i].RHS[j][k])
						k++
					}
					j++
				}
			}
		}
		if same {
			i++
		}
	}
	return same
}

// Determines whether the two given rule symbols
// are the equivalent.
func isSameSymbol(s1 RHSRuleSymbol, s2 RHSRuleSymbol) bool {
	same := true
	if reflect.TypeOf(s1).String() != reflect.TypeOf(s2).String() ||
		s1.Name() != s2.Name() {
		same = false
	} else {
		paramsIface := s1.Params()
		if params, isNtParams := paramsIface.(*NtRHSParams); isNtParams {
			g2ParamsIface := s2.Params()
			g2Params, isG2NtParams := g2ParamsIface.(*NtRHSParams)
			if !isG2NtParams {
				same = false
			} else if g2Params != nil && params != nil {
				if g2Params.Optional != (*bool)(nil) && params.Optional != (*bool)(nil) &&
					*g2Params.Optional != *params.Optional {
					same = false
				} else {
					if len(params.Conditions) != len(g2Params.Conditions) {
						same = false
					} else {
						l := 0
						for same && l < len(params.Conditions) {
							if params.Conditions[l] != g2Params.Conditions[l] {
								same = false
							}
							l++
						}
						if same {
							if len(params.Passthrough) != len(g2Params.Passthrough) {
								same = false
							} else {
								l := 0
								for same && l < len(params.Passthrough) {
									if params.Passthrough[l] != g2Params.Passthrough[l] {
										same = false
									}
									l++
								}
							}
						}
					}
				}
			}
		} else if params, isTParams := paramsIface.(*TRHSParams); isTParams {
			g2ParamsIface := s2.Params()
			g2Params, isg2TParams := g2ParamsIface.(*TRHSParams)
			if !isg2TParams {
				same = false
			} else if g2Params != nil && params != nil {
				if len(params.Conditions) != len(g2Params.Conditions) {
					same = false
				} else {
					l := 0
					for same && l < len(params.Conditions) {
						if params.Conditions[l] != g2Params.Conditions[l] {
							same = false
						}
						l++
					}
				}
			}
		} else if params, isCParams := paramsIface.(*CRHSParams); isCParams {
			g2ParamsIface := s2.Params()
			g2Params, isg2TParams := g2ParamsIface.(*CRHSParams)
			if !isg2TParams {
				same = false
			} else if params != nil && g2Params != nil {
				if len(params.Conditions) != len(g2Params.Conditions) {
					same = false
				} else {
					l := 0
					for same && l < len(params.Conditions) {
						if params.Conditions[l] != g2Params.Conditions[l] {
							same = false
						}
						l++
					}
				}
			}
		} else if params, isLAParams := paramsIface.(*LaRHSParams); isLAParams {
			g2ParamsIface := s2.Params()
			g2Params, isg2TParams := g2ParamsIface.(*LaRHSParams)
			if !isg2TParams {
				same = false
			} else if params != nil && g2Params != nil {
				if len(params.Exclude) != len(g2Params.Exclude) {
					same = false
				} else {
					l := 0
					for same && l < len(params.Exclude) {
						rule := params.Exclude[l]
						m := 0
						for same && m < len(rule) {
							symbol := rule[m]
							// Generally there are no parameters in exclusion rule listings (Not in ECMAScript)
							// so we can just check against the name of the symbol.
							if symbol.Name() != g2Params.Exclude[l][m].Name() {
								same = false
							}
							m++
						}
						l++
					}
				}
			}
		} else {
			ruleSymbolIface := s1
			conditional, isCond := ruleSymbolIface.(*ConditionalRHSRuleSymbol)
			if isCond {
				g2ruleSymbolIface := s2
				g2Conditional, isG2Cond := g2ruleSymbolIface.(*ConditionalRHSRuleSymbol)
				if !isG2Cond || (g2Conditional == nil && conditional != nil) ||
					(g2Conditional != nil && conditional == nil) {
					same = false
				} else {
					if len(g2Conditional.Parts) != len(conditional.Parts) {
						same = false
					} else {
						i := 0
						for same && i < len(conditional.Parts) {
							same = isSameSymbol(conditional.Parts[i], g2Conditional.Parts[i])
							i++
						}
						if same {
							if (conditional.params != nil && g2Conditional.params == nil) ||
								(g2Conditional.params != nil && conditional.params == nil) {
								same = false
							} else if conditional.params != nil && g2Conditional.params != nil {
								if len(conditional.params.Conditions) != len(g2Conditional.params.Conditions) {
									same = false
								} else {
									for same && i < len(conditional.params.Conditions) {
										if conditional.params.Conditions[i] != g2Conditional.params.Conditions[i] {
											same = false
										}
										i++
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return same
}

func assertSame(t *testing.T, input Grammar, expected Grammar) {
	if !isSameGrammar(input, expected) {
		t.Errorf("Expected grammar \n%v but got grammar \n%v", sprintGrammar(expected), sprintGrammar(input))
	}
}
