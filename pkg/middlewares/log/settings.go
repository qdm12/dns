package log

import "strings"

type Settings struct {
	LogRequests  bool
	LogResponses bool
}

func (s *Settings) String() string {
	const (
		subSection = " |--"
		indent     = "    " // used if lines already contain the subSection
	)
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *Settings) Lines(indent, subSection string) (lines []string) {
	if !s.LogRequests && !s.LogResponses {
		return []string{subSection + "Status: disabled"}
	}

	if s.LogRequests {
		lines = append(lines, subSection+"Log requests: on")
	}

	if s.LogResponses {
		lines = append(lines, subSection+"Log responses: on")
	}

	return lines
}
