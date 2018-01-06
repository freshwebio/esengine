package grammar

import (
	"fmt"
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// ExtractConditionalPartRules deals with extracting conditional
// parts which make up a right hand side rule which is
// prefixed with one or more conditions.
func ExtractConditionalPartRules(list []interface{}) ([]RHSRuleSymbol, error) {
	var parts []RHSRuleSymbol
	for _, symbol := range list {
		symbolStr, symbolIsStr := symbol.(string)
		if symbolIsStr {
			matched, err := regexp.MatchString("^<\\w+>$", symbolStr)
			if err != nil {
				return parts, err
			}
			if matched {
				parts = append(parts, &NonTerminalRHSRuleSymbol{
					name: strings.TrimSuffix(strings.TrimPrefix(symbolStr, "<"), ">"),
				})
			} else if symbolStr == "<!LineTerminator!>" {
				parts = append(parts, &ExcludeRHSRuleSymbol{
					name: "LineTerminator",
				})
			} else {
				// If not a non-terminal for conditional parts then it
				// is a terminal symbol.
				parts = append(parts, &TerminalRHSRuleSymbol{
					name: symbolStr,
				})
			}
		} else {
			rulePartMapSlice, rulePartIsMapSlice := symbol.(yaml.MapSlice)
			if rulePartIsMapSlice {
				for _, v := range rulePartMapSlice {
					vMap, valueIsMapSlice := v.Value.(yaml.MapSlice)
					if valueIsMapSlice {
						name, isNameString := v.Key.(string)
						if isNameString {
							matched, err := regexp.MatchString("^<\\w+>$", name)
							if err != nil {
								return parts, err
							}
							if matched {
								paramIface := get(vMap, "params")
								paramMap, isParamMapSlice := paramIface.(yaml.MapSlice)
								if paramIface != nil && isParamMapSlice {
									params := &NtRHSParams{}
									conditions := get(paramMap, "conditions")
									pt := get(paramMap, "passthrough")
									optional := get(paramMap, "optional")
									if conditions != nil {
										condSlice := conditions.([]interface{})
										for _, condParam := range condSlice {
											condParamStr := condParam.(string)
											params.Conditions = append(params.Conditions, condParamStr)
										}
									}
									if pt != nil {
										ptSlice := pt.([]interface{})
										for _, ptParam := range ptSlice {
											ptParamStr := ptParam.(string)
											params.Passthrough = append(params.Passthrough, ptParamStr)
										}
									}
									if optional != nil {
										optionalBool := optional.(bool)
										params.Optional = &optionalBool
									}
									parts = append(parts, &NonTerminalRHSRuleSymbol{
										name:   strings.TrimSuffix(strings.TrimPrefix(name, "<"), ">"),
										params: params,
									})
								}
							} else {
								// It is a terminal symbol with parameters.
								paramIface := get(vMap, "params")
								paramMap, isParamMapSlice := paramIface.(yaml.MapSlice)
								if paramIface != nil && isParamMapSlice {
									params := &TRHSParams{}
									conditions := get(paramMap, "conditions")
									if conditions != nil {
										condSlice := conditions.([]interface{})
										for _, condParam := range condSlice {
											condParamStr := condParam.(string)
											params.Conditions = append(params.Conditions, condParamStr)
										}
									}
									parts = append(parts, &TerminalRHSRuleSymbol{
										name:   name,
										params: params,
									})
								}
							}
						}
					}
				}
			}
		}
	}
	return parts, nil
}

// ExtractLookaheadExclusions Deals with extracting
// the rules utilised by lookahead exclusions.
func ExtractLookaheadExclusions(list []interface{}) [][]RHSRuleSymbol {
	var rules [][]RHSRuleSymbol
	for _, rhs := range list {
		ruleParts := []RHSRuleSymbol{}
		symbols := rhs.([]interface{})
		for _, s := range symbols {
			symbolStr := s.(string)
			if symbolStr == "<!LineTerminator!>" {
				ruleParts = append(ruleParts, &ExcludeRHSRuleSymbol{
					name: "LineTerminator",
				})
			} else {
				ruleParts = append(ruleParts, &TerminalRHSRuleSymbol{
					name: symbolStr,
				})
			}
		}
		rules = append(rules, ruleParts)
	}
	return rules
}

// Prints the provided grammar to a string
// for debugging purposes.
func sprintGrammar(grammar Grammar) string {
	output := ""
	for _, prod := range grammar.Productions {
		params := "["
		for i, param := range prod.Params {
			params += param
			if i < len(prod.Params)-1 {
				params += ", "
			}
		}
		params += "]"
		output += prod.Name + params + ":\n"
		for _, rule := range prod.RHS {
			symbolNames := ""
			for _, symbol := range rule {
				symbolName := symbol.Name()
				if len(symbol.Name()) == 0 {
					symbolConditional, isCond := symbol.(*ConditionalRHSRuleSymbol)
					if isCond {
						symbolName = "["
						for i, condition := range symbolConditional.params.Conditions {
							if i == 0 {
								symbolName += condition
							} else {
								symbolName += ", " + condition
							}
						}
						symbolName += "]"
						for _, part := range symbolConditional.Parts {
							if _, isExclude := part.(*ExcludeRHSRuleSymbol); isExclude {
								symbolName += " [no " + part.Name() + " here]"
							} else {
								symbolName += " " + part.Name()
							}
						}
					} else {
						symbolLookahead, isLa := symbol.(*LookaheadRHSRuleSymbol)
						if isLa {
							symbolName = "[lookahead ∉ 〈 "
							for i, exclude := range symbolLookahead.params.Exclude {
								if i > 0 {
									symbolName += ", "
								}
								for j, excludeRule := range exclude {
									if j > 0 {
										symbolName += " "
									}
									if _, isExclude := excludeRule.(*ExcludeRHSRuleSymbol); isExclude {
										symbolName += "[no " + excludeRule.Name() + " here]"
									} else {
										symbolName += excludeRule.Name()
									}
								}
							}
							symbolName += " 〉]"
						}
					}
				}
				if _, isExclude := symbol.(*ExcludeRHSRuleSymbol); isExclude {
					symbolName = "[no " + symbol.Name() + " here]"
				}
				symbolNames += symbolName + " "
			}
			output += "    - " + symbolNames + "\n"
		}
		output += "\n"
	}
	return output
}

// Prints the provided grammar to std output
// for debugging purposes.
func printGrammar(grammar Grammar) {
	fmt.Print(sprintGrammar(grammar))
}
