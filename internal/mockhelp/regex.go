package mockhelp

import (
	"regexp"
)

func NewMatcherRegex(regex string) *MatcherRegex {
	return &MatcherRegex{
		regex: regexp.MustCompile(regex),
	}
}

type MatcherRegex struct {
	regex *regexp.Regexp
}

func (m *MatcherRegex) String() string {
	return "must match regex " + m.regex.String()
}

func (m *MatcherRegex) Matches(x interface{}) bool {
	s, ok := x.(string)
	if !ok {
		return false
	}

	return m.regex.MatchString(s)
}
