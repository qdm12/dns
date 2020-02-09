package models

import (
	"fmt"
	"net"
	"strings"
)

// ProviderData contains information for a DNS provider
type ProviderData struct {
	IPs  []net.IP
	Host Host
}

// Settings represents all the user settings for Unbound
type Settings struct {
	Providers             []Provider
	PrivateAddresses      []string
	ListeningPort         uint16
	VerbosityLevel        uint8
	VerbosityDetailsLevel uint8
	ValidationLogLevel    uint8
	Caching               bool
	BlockMalicious        bool
	BlockSurveillance     bool
	BlockAds              bool
	BlockedHostnames      []string
	BlockedIPs            []string
	AllowedHostnames      []string
}

func (s *Settings) String() string {
	caching, blockMalicious, blockSurveillance, blockAds := "disabed", "disabed", "disabed", "disabed"
	if s.Caching {
		caching = "enabled"
	}
	if s.BlockMalicious {
		blockMalicious = "enabled"
	}
	if s.BlockSurveillance {
		blockSurveillance = "enabled"
	}
	if s.BlockAds {
		blockAds = "enabled"
	}
	var providersStr []string
	for _, provider := range s.Providers {
		providersStr = append(providersStr, string(provider))
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
	}
	return strings.Join(settingsList, "\n")
}
