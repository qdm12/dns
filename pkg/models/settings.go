package models

import (
	"strconv"
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

func (s *Settings) Lines() (lines []string) {
	const (
		disabled = "disabled"
		enabled  = "enabled"
	)
	const prefix = " |--"

	lines = append(lines, "DNS over TLS provider:")
	for _, provider := range s.Providers {
		lines = append(lines, prefix+provider)
	}

	lines = append(lines,
		"Listening port: "+strconv.Itoa(int(s.ListeningPort)))

	caching := disabled
	if s.Caching {
		caching = enabled
	}
	lines = append(lines,
		"Caching: "+caching)

	ipv4 := disabled
	if s.IPv4 {
		ipv4 = enabled
	}
	lines = append(lines,
		"IPv4 resolution: "+ipv4)

	ipv6 := disabled
	if s.IPv6 {
		ipv6 = enabled
	}
	lines = append(lines,
		"IPv6 resolution: "+ipv6)

	lines = append(lines,
		"Verbosity level: "+strconv.Itoa(int(s.VerbosityLevel))+"/5")

	lines = append(lines,
		"Verbosity details level: "+strconv.Itoa(int(s.VerbosityDetailsLevel))+"/4")

	lines = append(lines,
		"Validation log level: "+strconv.Itoa(int(s.ValidationLogLevel))+"/2")

	lines = append(lines, "Blocked hostnames:")
	for _, hostname := range s.BlockedHostnames {
		lines = append(lines, prefix+hostname)
	}

	lines = append(lines, "Blocked IP addresses:")
	for _, ip := range s.BlockedIPs {
		lines = append(lines, prefix+ip)
	}

	lines = append(lines, "Allowed hostnames:")
	for _, hostname := range s.AllowedHostnames {
		lines = append(lines, prefix+hostname)
	}

	return lines
}
