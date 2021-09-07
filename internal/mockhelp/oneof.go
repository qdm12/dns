package mockhelp

import (
	"strings"
)

func NewMatcherOneOf(possibilities ...string) *MatcherOneOf {
	return &MatcherOneOf{
		possibilities: possibilities,
	}
}

type MatcherOneOf struct {
	possibilities []string
}

func (m *MatcherOneOf) String() string {
	return "must be one of: " + strings.Join(m.possibilities, ", ")
}

func (m *MatcherOneOf) Matches(x interface{}) bool {
	s, ok := x.(string)
	if !ok {
		return false
	}

	for _, possibility := range m.possibilities {
		if s == possibility {
			return true
		}
	}

	return false
}
