package models

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/qdm12/dns/pkg/models"
)

// ProviderData contains information for a DNS provider.
type ProviderData struct {
	IPs          []net.IP
	Host         Host
	SupportsIPv6 bool
}

type Settings struct { //nolint:maligned
	Unbound           models.Settings
	Username          string
	Puid, Pgid        int
	BlockMalicious    bool
	BlockAds          bool
	BlockSurveillance bool
	CheckUnbound      bool
	UpdatePeriod      time.Duration
}

func (s *Settings) String() string {
	const (
		disabled = "disabled"
		enabled  = "enabled"
	)
	blockMalicious, blockSurveillance, blockAds,
		checkUnbound, update :=
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
	if s.CheckUnbound {
		checkUnbound = enabled
	}
	if s.UpdatePeriod > 0 {
		update = fmt.Sprintf("every %s", s.UpdatePeriod)
	}
	settingsList := []string{
		"Unbound settings:\n|--" + strings.Join(s.Unbound.Lines(), "\n|--"),
		"Username: " + s.Username,
		"Process UID: " + strconv.Itoa(s.Puid),
		"Process GID: " + strconv.Itoa(s.Pgid),
		"Block malicious: " + blockMalicious,
		"Block ads: " + blockAds,
		"Block surveillance: " + blockSurveillance,
		"Check Unbound: " + checkUnbound,
		"Update: " + update,
	}
	return strings.Join(settingsList, "\n")
}
