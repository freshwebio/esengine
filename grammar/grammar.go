package grammar

import (
	"regexp"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Grammar provides the data structure
// for the ECMAScript Syntactic Grammar.
type Grammar struct {
	Productions []*Production
}

// UnmarshalYAML deals with converting YAML input data
// into the data structures used in this application in order to
// generate our parser components.
func (g *Grammar) UnmarshalYAML(unmarshal func(interface{}) error) error {
	target := &yaml.MapSlice{}
	err := unmarshal(target)
	if err != nil {
		return err
	}
	for _, p := range *target {
		matched, err := regexp.MatchString("^<\\w+>$", p.Key.(string))
		if err != nil {
			return err
		}
		if matched {
			pMap := p.Value.(yaml.MapSlice)
			params := get(pMap, "params")
			rhs := get(pMap, "rhs")
			prod := &Production{}
			prod.Name = strings.TrimSuffix(strings.TrimPrefix(p.Key.(string), "<"), ">")
			if params != nil {
				paramsIface := params.([]interface{})
				for _, param := range paramsIface {
					prod.Params = append(prod.Params, param.(string))
				}
			}
			if rhs != nil {
				err = prod.extractRHSRuleSymbols(rhs)
				if err != nil {
					return err
				}
			}
			g.Productions = append(g.Productions, prod)
			//fmt.Printf("%+s\n\n", prod)
		}
	}
	return nil
}

// Retrieves the item with the provided key in
// the specified map slice or nil
func get(mapSlc yaml.MapSlice, key string) interface{} {
	i := 0
	found := false
	for !found && i < len(mapSlc) {
		if mapSlc[i].Key == key {
			found = true
		} else {
			i++
		}
	}
	if i < len(mapSlc) {
		return mapSlc[i].Value
	}
	return nil
}
