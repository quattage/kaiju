/******************************************************************************/
/* selector.go                                                                */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rules

import "strings"

type RuleState = int

const (
	ReadingTag = iota
	ReadingId
	ReadingClass
	ReadingDescendant
	ReadingChild
	ReadingSibling
	ReadingAdjacent
	ReadingCondition
	ReadingConditionAssignment
	ReadingPseudo
	ReadingPseudoFunction
	ReadingProperty
	ReadingPropertyValue
	ReadingPropertyFunction
)

type SelectorPart struct {
	Name       string
	Args       []string
	SelectType RuleState
}

type Selector struct {
	Parts []SelectorPart
}

type MediaQuery struct {
	Key   string
	Value string
}

type SelectorGroup struct {
	Selectors  []Selector
	Rules      []Rule
	MediaQuery MediaQuery
}

func (r *SelectorPart) String() string {
	var output strings.Builder
	output.Grow(len(r.Args) * 5)
	output.WriteString(r.Name)
	output.WriteString(": [")
	for x := 0; x < len(r.Args); x++ {
		arg := r.Args[x]
		output.WriteString(arg)
		output.WriteString(", ")
	}
	return output.String()[:len(output.String())-2] + "]"
}

func (m *MediaQuery) IsValid() bool { return m.Key != "" }

func (m *MediaQuery) Clear() {
	m.Key = ""
	m.Value = ""
}

func (s *SelectorGroup) AddRule(r Rule) {
	s.Rules = append(s.Rules, r)
}
