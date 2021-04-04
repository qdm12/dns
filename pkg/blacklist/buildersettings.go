package blacklist

import (
	"net"
	"strconv"
	"strings"
)

type BuilderSettings struct {
	BlockMalicious    bool
	BlockAds          bool
	BlockSurveillance bool
	AllowedHosts      []string
	AddBlockedHosts   []string
	AddBlockedIPs     []net.IP
	AddBlockedIPNets  []*net.IPNet
}

func (s *BuilderSettings) String() string {
	const (
		subSection = " |--"
		indent     = "    " // used if lines already contain the subSection
	)
	return strings.Join(s.Lines(indent, subSection), "\n")
}

func (s *BuilderSettings) Lines(indent, subSection string) (lines []string) {
	const no = "no"
	const yes = "yes"
	blockMalicious, blockSurveillance, blockAds := no, no, no
	if s.BlockMalicious {
		blockMalicious = yes
	}
	if s.BlockSurveillance {
		blockSurveillance = yes
	}
	if s.BlockAds {
		blockAds = yes
	}
	lines = append(lines, subSection+"Block malicious: "+blockMalicious)
	lines = append(lines, subSection+"Block ads: "+blockAds)
	lines = append(lines, subSection+"Block surveillance: "+blockSurveillance)

	if len(s.AllowedHosts) > 0 {
		lines = append(lines, subSection+"Additional hostnames blocked: "+
			strconv.Itoa(len(s.AllowedHosts)))
	}

	if len(s.AddBlockedIPs) > 0 {
		lines = append(lines, subSection+"Additional IP addresses blocked: "+
			strconv.Itoa(len(s.AddBlockedIPs)))
	}

	if len(s.AddBlockedIPNets) > 0 {
		lines = append(lines, subSection+"Additional IP networks blocked: "+
			strconv.Itoa(len(s.AddBlockedIPNets)))
	}

	return lines
}
