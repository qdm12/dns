package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qdm12/dns/pkg/models"
)

type Settings struct {
	Unbound           models.Settings
	Username          string
	Puid, Pgid        int
	BlockMalicious    bool
	BlockAds          bool
	BlockSurveillance bool
	CheckDNS          bool
	UpdatePeriod      time.Duration
}

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
	lines = append(lines, subSection+"Username: "+s.Username)
	lines = append(lines, subSection+"Process UID: "+strconv.Itoa(s.Puid))
	lines = append(lines, subSection+"Process GID: "+strconv.Itoa(s.Pgid))
	lines = append(lines, subSection+"Block malicious: "+blockMalicious)
	lines = append(lines, subSection+"Block ads: "+blockAds)
	lines = append(lines, subSection+"Block surveillance: "+blockSurveillance)
	lines = append(lines, subSection+"Check DNS: "+checkDNS)
	lines = append(lines, subSection+"Update: "+update)

	return lines
}
