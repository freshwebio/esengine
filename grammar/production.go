package grammar

import (
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Production provides a left hand side production
// of the syntactic grammar of ECMAScript.
type Production struct {
	Name   string
	Params []string
	RHS    [][]RHSRuleSymbol
}

// Deals with extracting right hand side rules from the given
// potential set of right hand side rules.
func (p *Production) extractRHSRuleSymbols(rhsIface interface{}) error {
	rhs, isType := rhsIface.([]interface{})
	if !isType {
		return ErrInvalidRightHandSide
	}
	for i, ruleIface := range rhs {
		p.RHS = append(p.RHS, []RHSRuleSymbol{})
		rule, isRule := ruleIface.([]interface{})
		if isRule {
			for _, part := range rule {
				name, isStr := part.(string)
				if isStr {
					// In the case the rule part is a string we only expect
					// a non-terminal, terminal or an exclusion.
					// This is down to the fact that only these types of
					// symbols/functions that do not require parameters.
					matched, err := regexp.MatchString("^<\\w+>$", name)
					if err != nil {
						return err
					}
					if matched {
						p.RHS[i] = append(p.RHS[i], &NonTerminalRHSRuleSymbol{
							name: strings.TrimSuffix(strings.TrimPrefix(name, "<"), ">"),
						})
					} else {
						matched, err = regexp.MatchString("^<\\!\\w+\\!>$", name)
						if err != nil {
							return err
						}
						if matched {
							p.RHS[i] = append(p.RHS[i], &ExcludeRHSRuleSymbol{
								name: strings.TrimSuffix(strings.TrimPrefix(name, "<!"), "!>"),
							})
						} else {
							p.RHS[i] = append(p.RHS[i], &TerminalRHSRuleSymbol{
								name: name,
							})
						}
					}
				} else {
					// Extract the right hand side rule as a map
					// we can extract parameters from.
					pMap, isMapSlice := part.(yaml.MapSlice)
					if isMapSlice {
						for _, v := range pMap {
							name, keyIsStr := v.Key.(string)
							vMap, valueIsMapSlice := v.Value.(yaml.MapSlice)
							if keyIsStr && valueIsMapSlice {
								err := p.extractMapRule(name, i, vMap)
								if err != nil {
									return err
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

// Deals with extracting a rule as a map
// in order to extract parameters for the rule.
func (p *Production) extractMapRule(name string, i int, v yaml.MapSlice) error {
	matched, err := regexp.MatchString("^<\\w+>$", name)
	if err != nil {
		return err
	}
	paramIface := get(v, "params")
	paramMapSlice, isParamMapSlice := paramIface.(yaml.MapSlice)
	if matched {
		if paramIface != nil && isParamMapSlice {
			params := &NtRHSParams{}
			pt := get(paramMapSlice, "passthrough")
			conditions := get(paramMapSlice, "conditions")
			optional := get(paramMapSlice, "optional")
			if pt != nil {
				ptSlice := pt.([]interface{})
				for _, ptParam := range ptSlice {
					ptParamStr := ptParam.(string)
					params.Passthrough = append(params.Passthrough, ptParamStr)
				}
			}
			if conditions != nil {
				condSlice := conditions.([]interface{})
				for _, condParam := range condSlice {
					condParamStr := condParam.(string)
					params.Conditions = append(params.Conditions, condParamStr)
				}
			}
			if optional != nil {
				optionalBool := optional.(bool)
				params.Optional = &optionalBool
			}
			p.RHS[i] = append(p.RHS[i], &NonTerminalRHSRuleSymbol{
				name:   strings.TrimSuffix(strings.TrimPrefix(name, "<"), ">"),
				params: params,
			})
		}
	} else {
		if name == "<*Conditional*>" {
			params := &CRHSParams{}
			conditions := get(paramMapSlice, "conditions")
			if conditions != nil {
				condSlice := conditions.([]interface{})
				for _, condParam := range condSlice {
					condParamStr := condParam.(string)
					params.Conditions = append(params.Conditions, condParamStr)
				}
			}
			partsIface := get(v, "parts")
			partsList, isPartsList := partsIface.([]interface{})
			parts := []RHSRuleSymbol{}
			if partsIface != nil && isPartsList {
				parts, err = ExtractConditionalPartRules(partsList)
				if err != nil {
					return err
				}
			}
			p.RHS[i] = append(p.RHS[i], &ConditionalRHSRuleSymbol{
				Parts:  parts,
				params: params,
			})
		} else {
			if name == "<*Lookahead*>" {
				params := &LaRHSParams{}
				params.Exclude = [][]RHSRuleSymbol{}
				if paramIface != nil && isParamMapSlice {
					exclusionsIface := get(paramMapSlice, "exclude")
					exclusions, isExclusionList := exclusionsIface.([]interface{})
					if exclusionsIface != nil && isExclusionList {
						params.Exclude = ExtractLookaheadExclusions(exclusions)
					}
				}
				p.RHS[i] = append(p.RHS[i], &LookaheadRHSRuleSymbol{
					params: params,
				})
			} else {
				// If we get here, it is a terminal symbol which has some parameters.
				params := &TRHSParams{}
				if paramIface != nil && isParamMapSlice {
					conditions := get(paramMapSlice, "conditions")
					if conditions != nil {
						condSlice := conditions.([]interface{})
						for _, condParam := range condSlice {
							condParamStr := condParam.(string)
							params.Conditions = append(params.Conditions, condParamStr)
						}
					}
				}
				p.RHS[i] = append(p.RHS[i], &TerminalRHSRuleSymbol{
					name:   name,
					params: params,
				})
			}
		}
	}
	return nil
}
