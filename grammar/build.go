package grammar

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

// Build deals with producing the symbols
// and the parse table for the provided grammar
// file in the specified file that will be a part of the package
// specified.
func Build(grammarFile string, format string, outputFile string, pkg string) {
	switch format {
	case "yaml":
		buildGrammarYaml(grammarFile, outputFile, pkg)
	}
}

func buildGrammarYaml(grammarFile string, outputFile string, pkg string) {
	//output := "package " + pkg + "\n\n"
	file, err := ioutil.ReadFile(grammarFile)
	if err != nil {
		errFmt := fmt.Errorf("failed to read the file %v containing"+
			" the grammar in the YAML format\nerror: %v\n", grammarFile, err)
		log.Fatal(errFmt)
	}
	grammar := &Grammar{}
	err = yaml.Unmarshal(file, grammar)
	if err != nil {
		errFmt := fmt.Errorf("failed to parse the yaml"+
			" representation of the grammar\nerror: %v\n", err)
		log.Fatal(errFmt)
	}
	LLkify(grammar)
	printGrammar(*grammar)
	/*output += generateGrammarOutput(grammar)
	// TODO: write to output file.
	fmt.Println("Grammar output:")
	fmt.Println()
	fmt.Print(output)*/
}

// Deals with generating the parse symbol and table (LL(k)) output
// as a string which can then be written to a file as source code.
func generateGrammarOutput(grammar *Grammar) string {
	output := "const (\n    _ Symbol = iota\n"
	ntOutput := ""
	tOutput := ""
	ntSymbolNo := 0
	tSymbolNo := 0
	readTerminals := []string{}
	for _, prod := range grammar.Productions {
		ntOutput += "    ntSy" + strconv.Itoa(ntSymbolNo) + " // " + prod.Name + "\n"
		ntSymbolNo++
		if len(prod.Params) > 0 {
			possibilities := []string{}
			n := len(prod.Params)
			// Holds temporary current combinations.
			// (never should exceed 3 with ECMAScript grammar)
			data := [3]string{}
			combineParams(prod.Params, data, &possibilities, 0, n-1, 0, 2)
			combineParams(prod.Params, data, &possibilities, 0, n-1, 0, 3)
			possibilities = append(prod.Params, possibilities...)
			for _, poss := range possibilities {
				ntOutput += "    ntSy" + strconv.Itoa(ntSymbolNo) + " // " + prod.Name + "_" + poss + "\n"
				ntSymbolNo++
			}
		}
		for _, ruleset := range prod.RHS {
			for _, rule := range ruleset {
				terminal, isTerminal := rule.(*TerminalRHSRuleSymbol)
				if isTerminal && !contains(readTerminals, terminal.name) {
					tOutput += "    tSy" + strconv.Itoa(tSymbolNo) + " // " + terminal.name + "\n"
					readTerminals = append(readTerminals, terminal.name)
					tSymbolNo++
				}
			}
		}
	}
	output += ntOutput + tOutput + ")\n"
	return output
}

func contains(haystack []string, needle string) bool {
	found := false
	i := 0
	for !found && i < len(haystack) {
		if haystack[i] == needle {
			found = true
		} else {
			i++
		}
	}
	return found
}

func containsRuleSymbol(haystack []RHSRuleSymbol, needle RHSRuleSymbol) bool {
	found := false
	i := 0
	for !found && i < len(haystack) {
		if haystack[i].Name() == needle.Name() {
			found = true
		} else {
			// Now check on a conditional rule symbol's first part.
			conditional, isCond := needle.(*ConditionalRHSRuleSymbol)
			if isCond && len(conditional.Parts) > 0 &&
				conditional.Parts[0].Name() == haystack[i].Name() {
				found = true
			}
			i++
		}
	}
	return found
}

func containsRule(haystack [][]RHSRuleSymbol, needle []RHSRuleSymbol) bool {
	match := false
	i := 0
	for !match && i < len(haystack) {
		rule := haystack[i]
		if len(needle) == len(rule) {
			// In the case a rule houses a conditional, then we only ever expect
			// that rule to contain a single symbol.
			isCurrentRule := true
			if len(rule) == 1 {
				if cond, isCond := rule[0].(*ConditionalRHSRuleSymbol); isCond {
					needleCond, isNeedleCond := needle[0].(*ConditionalRHSRuleSymbol)
					if !isNeedleCond {
						isCurrentRule = false
					} else {
						if len(cond.Parts) != len(needleCond.Parts) {
							isCurrentRule = false
						} else {
							if len(cond.params.Conditions) != len(cond.params.Conditions) {
								isCurrentRule = false
							} else {
								j := 0
								for isCurrentRule && j < len(cond.Parts) {
									if cond.Parts[j].Name() != needleCond.Parts[j].Name() {
										isCurrentRule = false
									}
									j++
								}
								// In the case there is still a possibility of a match
								// then ensure the parameter conditions are the same.
								for isCurrentRule && j < len(cond.params.Conditions) {
									if cond.params.Conditions[j] != needleCond.params.Conditions[j] {
										isCurrentRule = false
									}
									j++
								}
							}
						}
					}
				}
			}
			j := 0
			for isCurrentRule && j < len(rule) {
				symbol := rule[j]
				if symbol.Name() != needle[j].Name() {
					isCurrentRule = false
				}
				j++
			}
			match = isCurrentRule
		}
		i++
	}
	return match
}

func combineParams(params []string, data [3]string, combined *[]string, start int, end int, index int, max int) {
	if index == max {
		combination := ""
		for j := 0; j < max; j++ {
			if j == 0 {
				combination += data[j]
			} else {
				combination += "_" + data[j]
			}
		}
		*combined = append(*combined, combination)
		return
	}
	for i := start; i <= end && end-i+1 >= max-index; i++ {
		data[index] = params[i]
		combineParams(params, data, combined, i+1, end, index+1, max)
	}
}

// LLkify deals with left-factoring and applying
// left recursion removal to make the grammar compatible with
// the ll(k) parsing algorithm.
func LLkify(grammar *Grammar) {
	EliminateLeftRecursion(grammar)
	LeftFactor(grammar)
}

// EliminateLeftRecursion deals with eliminating left recursion
// using the rule:
// A -> β₁A' | β₂A' | ... βⁿA'
// A' -> α₁A' | α₂A' | ... αⁿA'
// Then this goes on to handle further derivations of left recursion
// that are nested such as the grammar which follows.
// A -> C | e
// C -> A | bc
func EliminateLeftRecursion(grammar *Grammar) {
	newProductions := []*Production{}
	prevProductions := []*Production{}
	for _, prod := range grammar.Productions {
		// Removes previous productions from our current production.
		prod = removePrevProductions(prod, prevProductions)
		alphas := [][]RHSRuleSymbol{}
		betas := [][]RHSRuleSymbol{}
		for _, rule := range prod.RHS {
			if len(rule) > 0 {
				// If first symbol is a left recursion
				// then extract the alpha part of the rule.
				if rule[0].Name() == prod.Name {
					alphas = append(alphas, rule[1:])
				} else {
					// This is a non left recursive symbol
					// so make it a beta.
					betas = append(betas, rule)
				}
			}
		}
		// If there are alphas then we know we need to do some left
		// recursion elimation.
		if len(alphas) > 0 {
			newProdA := &Production{}
			newProdA.Name = prod.Name
			newProdA.Params = prod.Params
			newProdAPrime := &Production{}
			newProdAPrime.Name = prod.Name + "'"
			newProdAPrime.Params = prod.Params
			aPrimeRule := &NonTerminalRHSRuleSymbol{
				name: newProdAPrime.Name,
				params: &NtRHSParams{
					Passthrough: prefix(prod.Params, "?"),
				},
			}
			for _, beta := range betas {
				betaAPrime := append(beta, aPrimeRule)
				newProdA.RHS = append(newProdA.RHS, betaAPrime)
			}
			for _, alpha := range alphas {
				alphaAPrime := append(alpha, aPrimeRule)
				newProdAPrime.RHS = append(newProdAPrime.RHS, alphaAPrime)
			}
			newProdAPrime.RHS = append(newProdAPrime.RHS, []RHSRuleSymbol{&TerminalRHSRuleSymbol{
				name: "[empty]",
			}})
			newProductions = append(newProductions, newProdA, newProdAPrime)
		} else {
			newProductions = append(newProductions, prod)
		}
		prevProductions = append(prevProductions, prod)

	}
	grammar.Productions = newProductions
}

// Deals with removing previous productions from the current production
// for the sake of derivative left recursion removal by applying them and combining
// the right-hand side of a previous production with the provided production's
// right-hand side.
func removePrevProductions(prod *Production, prevProds []*Production) *Production {
	newProd := prod
	if len(prevProds) > 0 {
		newProd = &Production{
			Name:   prod.Name,
			Params: prod.Params,
		}
		for _, rule := range prod.RHS {
			// List of rules which are a result of applying
			// previous productions.
			prevAppliedRules := [][]RHSRuleSymbol{}
			for _, prev := range prevProds {
				if len(rule) > 0 && rule[0].Name() == prev.Name {
					delta := rule[1:]
					// Remove productions of the form A -> ε before attempting to iterate over the set of rules.
					if len(prev.RHS) == 1 && len(prev.RHS[0]) == 1 && prev.RHS[0][0].Name() == "[empty]" {
						if len(rule) == 1 {
							// If rule is epsilon (Contains a single [empty] rule symbol)
							// then we will replace our non-terminal with epsilon.
							epsilon := []RHSRuleSymbol{&TerminalRHSRuleSymbol{
								name: "[empty]",
							}}
							if !ruleExistsInAny(epsilon, prevAppliedRules, prod.RHS) {
								prevAppliedRules = append(prevAppliedRules, epsilon)
							}
						} else {
							// If the rule is prefixed by our specified non-terminal then we will simply remove it
							// as epsilon followed by any other symbol would be the equivalent to
							// an empty string.
						}
					} else {
						// Apply the previous production and we replace
						// it in the current production with it's right hand side.
						for _, prevRule := range prev.RHS {
							// due to γ ≠ ε we will disregard any right-hand side rules
							// which are empty or only contain the [empty] terminal.
							if len(prevRule) > 1 || (len(prevRule) == 1 && prevRule[0].Name() != "[empty]") {
								// Append delta to the end of our new production.
								prevRule = append(prevRule, delta...)
								// Also, we'll avoid duplicates by checking at this point.
								if !ruleExistsInAny(prevRule, prevAppliedRules, prod.RHS) {
									prevAppliedRules = append(prevAppliedRules, prevRule)
								}
							}
						}
					}
				}
			}
			// No previously applied rules means there are no references to previous
			// productions in the given rule and we can just add the rule to our new
			// production as is.
			if len(prevAppliedRules) == 0 {
				newProd.RHS = append(newProd.RHS, rule)
			} else {
				newProd.RHS = append(newProd.RHS, prevAppliedRules...)
			}
		}
	}
	return newProd
}

// Determines whether the provided rule exists in any of the provided
// right-hand side rule sets.
func ruleExistsInAny(rule []RHSRuleSymbol, ruleSets ...[][]RHSRuleSymbol) bool {
	found := false
	i := 0
	for !found && i < len(ruleSets) {
		j := 0
		for !found && j < len(ruleSets[i]) {
			k := 0
			for !found && k < len(ruleSets[i][j]) {
				current := ruleSets[i][j]
				l := 0
				match := true
				if len(rule) == len(current) {
					for match && l < len(current) {
						if rule[l] != current[l] {
							match = false
						}
						l++
					}
				}
				found = match
				k++
			}
			j++
		}
		i++
	}
	return found
}

// LeftFactor deals with left-factoring productions
// in the given grammar using the following rule:
// replace A -> αβ₁ | ... αβⁿ | γ
// by
// A -> αA' | γ
// A' -> β₁ | ... | βⁿ
func LeftFactor(grammar *Grammar) {
	productions := []*Production{}
	for _, prod := range grammar.Productions {
		newProductions, newRules := leftFactorRules(prod.RHS, prod.Name, prod.Params)
		if len(newRules) > 0 {
			prod.RHS = newRules
		}
		productions = append(productions, prod)
		if len(newProductions) > 0 {
			productions = append(productions, newProductions...)
		}
	}
	grammar.Productions = productions
}

// Deals with left-factoring a given set of rules and generating a new set of productions
// in the case left-factoring is needed.
func leftFactorRules(rules [][]RHSRuleSymbol, prodName string, prodParams []string) ([]*Production, [][]RHSRuleSymbol) {
	// Holds rule symbols which are the start of more than one rule.
	var alphas []RHSRuleSymbol
	alphaBetaMap := make(map[string][][]RHSRuleSymbol)
	alphaGammaMap := make(map[string][][]RHSRuleSymbol)
	var newRules [][]RHSRuleSymbol
	var newProductions []*Production
	for i, rule := range rules {
		var betas [][]RHSRuleSymbol
		var gammas [][]RHSRuleSymbol
		j := 0
		leftRepeat := false
		for j < len(rules) {
			if j != i && isFirstSymbolSame(rule, rules[j]) {
				leftRepeat = true
				// To allow for empty as a beta to be used in production left-factoring
				//we will add empty (epsilon) symbols to betas where a rule is made of only alpha.
				beta := []RHSRuleSymbol{&TerminalRHSRuleSymbol{name: "[empty]"}}
				if len(rules[j]) > 1 {
					beta = rules[j][1:]
				} else if len(rules[j]) == 1 {
					conditional, isCond := rules[j][0].(*ConditionalRHSRuleSymbol)
					if isCond {
						if len(conditional.Parts) > 1 {
							beta = []RHSRuleSymbol{
								&ConditionalRHSRuleSymbol{
									params: conditional.params,
									Parts:  conditional.Parts[1:],
								},
							}
						}
					}
				}
				betas = append(betas, beta)
			} else if !isFirstSymbolSame(rule, rules[j]) {
				gammas = append(gammas, rules[j])
			}
			j++
		}
		if leftRepeat && !containsRuleSymbol(alphas, rule[0]) {
			alphas = append(alphas, rule[0])
			primaryRuleBeta := []RHSRuleSymbol{&TerminalRHSRuleSymbol{
				name: "[empty]",
			}}
			if len(rule) > 1 {
				primaryRuleBeta = rule[1:]
			}
			if len(primaryRuleBeta) == 1 && primaryRuleBeta[0].Name() == "[empty]" {
				alphaBetaMap[rule[0].Name()] = append(betas, primaryRuleBeta)
			} else {
				alphaBetaMap[rule[0].Name()] = append([][]RHSRuleSymbol{primaryRuleBeta}, betas...)
			}
			alphaGammaMap[rule[0].Name()] = gammas
		}
	}
	// Now handle each alpha one at a time in producing our new rules and productions.
	if len(alphas) > 0 {
		for i := len(alphas) - 1; i >= 0; i-- {
			alpha := alphas[i]
			prodPrimeRule := &NonTerminalRHSRuleSymbol{
				name: prodName + "A" + strconv.Itoa(i),
				params: &NtRHSParams{
					Passthrough: prefix(prodParams, "?"),
				},
			}
			newRules = append(
				[][]RHSRuleSymbol{[]RHSRuleSymbol{alpha, prodPrimeRule}},
				newRules...,
			)
			for _, gamma := range alphaGammaMap[alpha.Name()] {
				// Only in the case our gamma's first rule
				// is not an alpha and was not a part of alpha[i+1]
				// we will add it to the new set of rules.
				if !containsRuleSymbol(alphas, gamma[0]) {
					if (i+1 < len(alphas) && !containsRule(alphaGammaMap[alphas[i+1].Name()], gamma)) ||
						i+1 >= len(alphas) {
						newRules = append(newRules, gamma)
					}
				}
			}
			prodPrime := &Production{
				Name:   prodName + "A" + strconv.Itoa(i),
				Params: prodParams,
			}
			for _, beta := range alphaBetaMap[alpha.Name()] {
				prodPrime.RHS = append(prodPrime.RHS, beta)
			}
			newProductions = append([]*Production{prodPrime}, newProductions...)

			// Now for A'.
			furtherPrimeProductions, newProdPrimeRules := leftFactorRules(prodPrime.RHS, prodPrime.Name, prodPrime.Params)
			if len(newProdPrimeRules) > 0 {
				prodPrime.RHS = newProdPrimeRules
			}
			newProductions = append(newProductions, furtherPrimeProductions...)
		}
	}
	return newProductions, newRules
}

// Adds the given prefix to each string in the provided list.
func prefix(strings []string, prefix string) []string {
	prefixed := []string{}
	for _, str := range strings {
		prefixed = append(prefixed, prefix+str)
	}
	return prefixed
}

// Determines whether the first symbol in the given rule sets
// are the same.
func isFirstSymbolSame(r1 []RHSRuleSymbol, r2 []RHSRuleSymbol) bool {
	r1FirstSymbolName := getSymbolName(r1[0])
	r2FirstSymbolName := getSymbolName(r2[0])
	return r1FirstSymbolName == r2FirstSymbolName
}

// Retrieves either the symbol name or in the case
// the symbol is a wrapper that contains symbols then
// the name of the first symbol contained.
func getSymbolName(s RHSRuleSymbol) string {
	name := ""
	_, isTerminal := s.(*TerminalRHSRuleSymbol)
	isNonTerminal := false
	if !isTerminal {
		_, isNonTerminal = s.(*NonTerminalRHSRuleSymbol)
	}
	if isTerminal || isNonTerminal {
		name = s.Name()
	} else {
		conditional, isCond := s.(*ConditionalRHSRuleSymbol)
		if isCond {
			if len(conditional.Parts) > 0 {
				name = conditional.Parts[0].Name()
			}
		}
	}
	return name
}
