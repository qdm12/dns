package env

import (
	"fmt"
	"net/netip"
	"os"
	"strings"
	"time"

	"github.com/qdm12/govalid/binary"
)

func envToCSV(envKey string) (values []string) {
	csv := os.Getenv(envKey)
	if csv == "" {
		return nil
	}
	return lowerAndSplit(csv)
}

func envToStringPtr(envKey string) (stringPtr *string) {
	s := os.Getenv(envKey)
	if s == "" {
		return nil
	}
	return &s
}

func envToBoolPtr(envKey string) (boolPtr *bool, err error) {
	s := os.Getenv(envKey)
	if s == "" {
		return nil, nil //nolint:nilnil
	}
	value, err := binary.Validate(s)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func envToDuration(envKey string) (d time.Duration, err error) {
	s := os.Getenv(envKey)
	if s == "" {
		return 0, nil
	}

	d, err = time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	return d, nil
}

func envToDurationPtr(envKey string) (d *time.Duration, err error) {
	s := os.Getenv(envKey)
	if s == "" {
		return nil, nil //nolint:nilnil
	}

	d = new(time.Duration)
	*d, err = time.ParseDuration(s)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func lowerAndSplit(csv string) (values []string) {
	csv = strings.ToLower(csv)
	return strings.Split(csv, ",")
}

func parseIPStrings(ipStrings []string) (ips []netip.Addr, err error) {
	ips = make([]netip.Addr, len(ipStrings))

	for i, ipString := range ipStrings {
		ips[i], err = netip.ParseAddr(ipString)
		if err != nil {
			return nil, fmt.Errorf("IP address string is not valid: %w", err)
		}
	}

	return ips, nil
}

func parseIPPrefixStrings(ipPrefixStrings []string) (ipPrefixes []netip.Prefix, err error) {
	ipPrefixes = make([]netip.Prefix, len(ipPrefixStrings))

	for i, ipPrefixString := range ipPrefixStrings {
		ipPrefixes[i], err = netip.ParsePrefix(ipPrefixString)
		if err != nil {
			return nil, fmt.Errorf("IP prefix CIDR string is not valid: %w", err)
		}
	}

	return ipPrefixes, nil
}
