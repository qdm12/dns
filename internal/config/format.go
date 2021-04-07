package config

import (
	"strings"
)

func (s *Settings) String() string {
	return strings.Join(s.Lines("   ", " |--"), "\n")
}

func (s *Settings) Lines(indent, subSection string) (lines []string) {
	lines = append(lines, subSection+"Upstream type: "+string(s.UpstreamType))

	switch s.UpstreamType {
	case DoT:
		lines = append(lines, subSection+"DoT settings:")
		for _, line := range s.DoT.Lines(indent, subSection) {
			lines = append(lines, indent+line)
		}
	case DoH:
		lines = append(lines, subSection+"DoH settings:")
		for _, line := range s.DoH.Lines(indent, subSection) {
			lines = append(lines, indent+line)
		}
	}

	lines = append(lines, subSection+"Blacklist settings:")
	for _, line := range s.Blacklist.Lines(indent, subSection) {
		lines = append(lines, indent+line)
	}

	const disabled, enabled = "disabled", "enabled"
	checkDNS := disabled
	if s.CheckDNS {
		checkDNS = enabled
	}
	lines = append(lines, subSection+"Check DNS: "+checkDNS)

	update := disabled
	if s.UpdatePeriod > 0 {
		update = "every " + s.UpdatePeriod.String()
	}
	lines = append(lines, subSection+"Update: "+update)

	lines = append(lines, subSection+"Log level: "+s.LogLevel.String())

	return lines
}
