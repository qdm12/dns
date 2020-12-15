package models

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// ProviderData contains information for a DNS provider.
type ProviderData struct {
	IPs          []net.IP
	Host         Host
	SupportsIPv6 bool
}

// Settings represents all the user settings for Unbound.
type Settings struct { //nolint:maligned
	Providers             []Provider
	ListeningPort         uint16
	Caching               bool
	IPv4                  bool
	IPv6                  bool
	VerbosityLevel        uint8
	VerbosityDetailsLevel uint8
	ValidationLogLevel    uint8
	BlockMalicious        bool
	BlockSurveillance     bool
	BlockAds              bool
	BlockedHostnames      []string
	BlockedIPs            []string
	AllowedHostnames      []string
	PrivateAddresses      []string
	CheckUnbound          bool
	UpdatePeriod          time.Duration
}

func (s *Settings) String() string {
	const (
		disabled = "disabled"
		enabled  = "enabled"
	)
	caching, blockMalicious, blockSurveillance, blockAds,
		checkUnbound, ipv4, ipv6, update :=
		disabled, disabled, disabled, disabled,
		disabled, disabled, disabled, disabled
	if s.Caching {
		caching = enabled
	}
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
	if s.IPv4 {
		ipv4 = enabled
	}
	if s.IPv6 {
		ipv6 = enabled
	}
	if s.UpdatePeriod > 0 {
		update = fmt.Sprintf("every %s", s.UpdatePeriod)
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
	privateAddresses := "Private addresses:"
	if len(s.PrivateAddresses) > 0 {
		privateAddresses += " \n |--" + strings.Join(s.PrivateAddresses, "\n |--")
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
		"Block malicious: " + blockMalicious,
		"Block surveillance: " + blockSurveillance,
		"Block ads: " + blockAds,
		blockedHostnames,
		blockedIPs,
		allowedHostnames,
		privateAddresses,
		"Check Unbound: " + checkUnbound,
		"Update: " + update,
	}
	return strings.Join(settingsList, "\n")
}
