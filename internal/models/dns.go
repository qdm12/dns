package models

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// ProviderData contains information for a DNS provider.
type ProviderData struct {
	IPs          []net.IP
	Host         Host
	SupportsIPv6 bool
}

type Settings struct { //nolint:maligned
	Unbound           UnboundSettings
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
		"Unbound settings:\n|--" + strings.Join(s.Unbound.lines(), "\n|--"),
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

// UnboundSettings represents all the user settings for Unbound.
type UnboundSettings struct { //nolint:maligned
	Providers             []Provider
	ListeningPort         uint16
	Caching               bool
	IPv4                  bool
	IPv6                  bool
	VerbosityLevel        uint8
	VerbosityDetailsLevel uint8
	ValidationLogLevel    uint8
	BlockedHostnames      []string
	BlockedIPs            []string
	AllowedHostnames      []string
}

func (s *UnboundSettings) lines() []string {
	const (
		disabled = "disabled"
		enabled  = "enabled"
	)
	caching, ipv4, ipv6 := disabled, disabled, disabled
	if s.Caching {
		caching = enabled
	}
	if s.IPv4 {
		ipv4 = enabled
	}
	if s.IPv6 {
		ipv6 = enabled
	}
	providersStr := make([]string, len(s.Providers))
	for i := range s.Providers {
		providersStr[i] = string(s.Providers[i])
	}
	blockedHostnames := "Blocked hostnames:"
	if len(s.BlockedHostnames) > 0 {
		blockedHostnames += " \n |--" + strings.Join(s.BlockedHostnames, "\n |--")
	}
	blockedIPs := "Blocked IP addresses:"
	if len(s.BlockedIPs) > 0 {
		blockedIPs += " \n |--" + strings.Join(s.BlockedIPs, "\n |--")
	}
	allowedHostnames := "Allowed hostnames:"
	if len(s.AllowedHostnames) > 0 {
		allowedHostnames += " \n |--" + strings.Join(s.AllowedHostnames, "\n |--")
	}
	settingsList := []string{
		"DNS over TLS provider:\n|--" + strings.Join(providersStr, "\n|--"),
		"Listening port: " + fmt.Sprintf("%d", s.ListeningPort),
		"Caching: " + caching,
		"IPv4 resolution: " + ipv4,
		"IPv6 resolution: " + ipv6,
		"Verbosity level: " + fmt.Sprintf("%d/5", s.VerbosityLevel),
		"Verbosity details level: " + fmt.Sprintf("%d/4", s.VerbosityDetailsLevel),
		"Validation log level: " + fmt.Sprintf("%d/2", s.ValidationLogLevel),
		blockedHostnames,
		blockedIPs,
		allowedHostnames,
	}
	return settingsList
}
