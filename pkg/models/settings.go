package models

import (
	"fmt"
	"strings"
)

// Settings represents all the user settings for Unbound.
type Settings struct { //nolint:maligned
	Providers             []string
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

func (s *Settings) Lines() []string {
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
		"DNS over TLS provider:\n|--" + strings.Join(s.Providers, "\n|--"),
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
