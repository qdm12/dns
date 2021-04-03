package config

import (
	"fmt"
	"strings"
)

func (s *Settings) String() string {
	return strings.Join(s.Lines("   ", " |--"), "\n")
}

func (s *Settings) Lines(indent, subSection string) (lines []string) {
	const (
		disabled = "disabled"
		enabled  = "enabled"
	)
	blockMalicious, blockSurveillance, blockAds,
		checkDNS, update :=
		disabled, disabled, disabled,
		disabled, disabled
	if s.BlockMalicious {
		blockMalicious = enabled
	}
	if s.BlockSurveillance {
		blockSurveillance = enabled
	}
	if s.BlockAds {
		blockAds = enabled
	}
	if s.CheckDNS {
		checkDNS = enabled
	}
	if s.UpdatePeriod > 0 {
		update = fmt.Sprintf("every %s", s.UpdatePeriod)
	}

	lines = append(lines, subSection+"Unbound settings:")
	for _, line := range s.Unbound.Lines() {
		lines = append(lines, indent+line)
	}
	lines = append(lines, subSection+"Block malicious: "+blockMalicious)
	lines = append(lines, subSection+"Block ads: "+blockAds)
	lines = append(lines, subSection+"Block surveillance: "+blockSurveillance)
	lines = append(lines, subSection+"Check DNS: "+checkDNS)
	lines = append(lines, subSection+"Update: "+update)

	return lines
}
