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
		"Allowed hostnames:\n |--" + strings.Join(s.AllowedHostnames, "\n |--"),
		"Private addresses:\n |--" + strings.Join(s.PrivateAddresses, "\n |--"),
	}
	return strings.Join(settingsList, "\n")
}
