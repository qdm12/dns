package env

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/qdm12/govalid/binary"
	"inet.af/netaddr"
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

var ErrIPStringNotValid = errors.New("IP address string is not valid")

func parseIPStrings(ipStrings []string) (ips []netaddr.IP, err error) {
	ips = make([]netaddr.IP, len(ipStrings))

	for i, ipString := range ipStrings {
		ips[i], err = netaddr.ParseIP(ipString)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrIPStringNotValid, err)
		}
	}

	return ips, nil
}

var ErrIPPrefixStringNotValid = errors.New("IP prefix CIDR string is not valid")

func parseIPPrefixStrings(ipPrefixStrings []string) (ipPrefixes []netaddr.IPPrefix, err error) {
	ipPrefixes = make([]netaddr.IPPrefix, len(ipPrefixStrings))

	for i, ipPrefixString := range ipPrefixStrings {
		ipPrefixes[i], err = netaddr.ParseIPPrefix(ipPrefixString)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrIPPrefixStringNotValid, err)
		}
	}

	return ipPrefixes, nil
}
