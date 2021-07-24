package config

import (
	"errors"
	"fmt"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/golibs/params"
	"inet.af/netaddr"
)

func getBlacklistSettings(reader *reader) (settings blacklist.BuilderSettings, err error) {
	settings.BlockMalicious, err = reader.env.OnOff("BLOCK_MALICIOUS", params.Default("on"))
	if err != nil {
		return settings, fmt.Errorf("environment variable BLOCK_MALICIOUS: %w", err)
	}
	settings.BlockSurveillance, err = reader.env.OnOff("BLOCK_SURVEILLANCE", params.Default("off"))
	if err != nil {
		return settings, fmt.Errorf("environment variable BLOCK_SURVEILLANCE: %w", err)
	}
	settings.BlockAds, err = reader.env.OnOff("BLOCK_ADS", params.Default("off"))
	if err != nil {
		return settings, fmt.Errorf("environment variable BLOCK_ADS: %w", err)
	}
	settings.AllowedHosts, err = getAllowedHostnames(reader)
	if err != nil {
		return settings, err
	}
	settings.AddBlockedHosts, err = getBlockedHostnames(reader)
	if err != nil {
		return settings, err
	}
	settings.AddBlockedIPs, err = getBlockedIPs(reader)
	if err != nil {
		return settings, err
	}
	settings.AddBlockedIPPrefixes, err = getBlockedIPPrefixes(reader)
	if err != nil {
		return settings, err
	}
	rebindingProtection, err := reader.env.OnOff("REBINDING_PROTECTION", params.Default("on"))
	if err != nil {
		return settings, fmt.Errorf("environment variable REBINDING_PROTECTION: %w", err)
	}
	if rebindingProtection {
		privateIPPrefixes, err := getPrivateIPPrefixes()
		if err != nil {
			return settings, err
		}
		settings.AddBlockedIPPrefixes = append(settings.AddBlockedIPPrefixes, privateIPPrefixes...)
	}

	return settings, nil
}

// getAllowedHostnames obtains a list of hostnames to unblock from block lists
// from the comma separated list for the environment variable UNBLOCK.
func getAllowedHostnames(reader *reader) (hostnames []string, err error) {
	hostnames, err = reader.env.CSV("ALLOWED_HOSTNAMES")
	if err != nil {
		return nil, fmt.Errorf("environment variable UNBLOCK: %w", err)
	}
	for _, hostname := range hostnames {
		if !reader.verifier.MatchHostname(hostname) {
			return nil, fmt.Errorf("%w: allowed hostname: %s", errHostnameInvalid, hostname)
		}
	}
	return hostnames, nil
}

var errHostnameInvalid = errors.New("hostname is invalid")

// getBlockedHostnames obtains a list of hostnames to block from the comma
// separated list for the environment variable BLOCK_HOSTNAMES.
func getBlockedHostnames(reader *reader) (hostnames []string, err error) {
	hostnames, err = reader.env.CSV("BLOCK_HOSTNAMES")
	if err != nil {
		return nil, fmt.Errorf("environment variable BLOCK_HOSTNAMES: %w", err)
	}
	for _, hostname := range hostnames {
		if !reader.verifier.MatchHostname(hostname) {
			return nil, fmt.Errorf("%w: blocked hostname: %s", errHostnameInvalid, hostname)
		}
	}
	return hostnames, nil
}

var errIPStringInvalid = errors.New("IP address string is invalid")

// getBlockedIPs obtains a list of IP addresses to block from
// the comma separated list for the environment variable BLOCK_IPS.
func getBlockedIPs(reader *reader) (ips []netaddr.IP, err error) {
	values, err := reader.env.CSV("BLOCK_IPS")
	if err != nil {
		return nil, fmt.Errorf("environment variable BLOCK_IPS: %w", err)
	}

	ips = make([]netaddr.IP, len(values))
	for _, value := range values {
		ip, err := netaddr.ParseIP(value)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errIPStringInvalid, err)
		}
		ips = append(ips, ip)
	}

	return ips, nil
}

var errBlockedIPPrefixInvalid = errors.New("blocked IP prefix CIDR is invalid")

// getBlockedIPPrefixes obtains a list of IP networks (CIDR notation) to block from
// the comma separated list for the environment variable BLOCK_CIDRS.
func getBlockedIPPrefixes(reader *reader) (ipPrefixes []netaddr.IPPrefix, err error) {
	values, err := reader.env.CSV("BLOCK_CIDRS")
	if err != nil {
		return nil, err
	}

	ipPrefixes = make([]netaddr.IPPrefix, len(values))
	for _, value := range values {
		ipPrefix, err := netaddr.ParseIPPrefix(value)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", errBlockedIPPrefixInvalid, err)
		}
		ipPrefixes = append(ipPrefixes, ipPrefix)
	}

	return ipPrefixes, nil
}

func getPrivateIPPrefixes() (privateIPPrefixes []netaddr.IPPrefix, err error) {
	privateCIDRs := []string{
		// IPv4 private addresses
		"127.0.0.1/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		// IPv6 private addresses
		"::1/128",
		"fc00::/7",
		"fe80::/10",
		// Private IPv4 addresses wrapped in IPv6
		"::ffff:7f00:1/104", // 127.0.0.1/8
		"::ffff:a00:0/104",  // 10.0.0.0/8
		"::ffff:ac10:0/108", // 172.16.0.0/12
		"::ffff:c0a8:0/112", // 192.168.0.0/16
		"::ffff:a9fe:0/112", // 169.254.0.0/16
	}
	privateIPPrefixes = make([]netaddr.IPPrefix, len(privateCIDRs))
	for i := range privateCIDRs {
		privateIPPrefixes[i], err = netaddr.ParseIPPrefix(privateCIDRs[i])
		if err != nil {
			return nil, err
		}
	}

	return privateIPPrefixes, nil
}
