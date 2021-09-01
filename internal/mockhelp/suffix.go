package mockhelp

import (
	"strings"
)

func NewMatcherStringSuffix(suffix string) *MatcherStringSuffix {
	return &MatcherStringSuffix{suffix: suffix}
}

type MatcherStringSuffix struct {
	suffix string
}

func (m *MatcherStringSuffix) String() string {
	return "string ending with: " + m.suffix
}
func (m *MatcherStringSuffix) Matches(x interface{}) bool {
	s, ok := x.(string)
	if !ok {
		return false
	}

	return strings.HasSuffix(s, m.suffix)
}
