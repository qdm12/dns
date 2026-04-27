package mockhelp

import (
	"slices"
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

func (m *MatcherOneOf) Matches(x any) bool {
	s, ok := x.(string)
	return ok && slices.Contains(m.possibilities, s)
}
