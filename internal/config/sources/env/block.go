package env

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
	"inet.af/netaddr"
)

func readBlock() (settings settings.Block, err error) {
	settings.BlockMalicious, err = envToBoolPtr("BLOCK_MALICIOUS")
	if err != nil {
		return settings, fmt.Errorf("environment variable BLOCK_MALICIOUS: %w", err)
	}

	settings.BlockSurveillance, err = envToBoolPtr("BLOCK_SURVEILLANCE")
	if err != nil {
		return settings, fmt.Errorf("environment variable BLOCK_SURVEILLANCE: %w", err)
	}

	settings.BlockAds, err = envToBoolPtr("BLOCK_ADS")
	if err != nil {
		return settings, fmt.Errorf("environment variable BLOCK_ADS: %w", err)
	}

	settings.RebindingProtection, err = envToBoolPtr("REBINDING_PROTECTION")
	if err != nil {
		return settings, fmt.Errorf("environment variable REBINDING_PROTECTION: %w", err)
	}

	settings.AllowedHosts = envToCSV("ALLOWED_HOSTNAMES")
	settings.AddBlockedHosts = envToCSV("BLOCK_HOSTNAMES")

	settings.AllowedIPs, err = getAllowedIPs()
	if err != nil {
		return settings, err
	}
	settings.AddBlockedIPs, err = getBlockedIPs()
	if err != nil {
		return settings, err
	}

	settings.AllowedIPPrefixes, err = getAllowedIPPrefixes()
	if err != nil {
		return settings, err
	}
	settings.AddBlockedIPPrefixes, err = getBlockedIPPrefixes()
	if err != nil {
		return settings, err
	}

	settings.RebindingProtection, err = envToBoolPtr("REBINDING_PROTECTION")
	if err != nil {
		return settings, fmt.Errorf("environment variable REBINDING_PROTECTION: %w", err)
	}

	return settings, nil
}

// getAllowedIPs obtains a list of IPs to unblock from block lists
// from the comma separated list for the environment variable ALLOWED_IPS.
func getAllowedIPs() (ips []netaddr.IP, err error) {
	ipStrings := envToCSV("ALLOWED_IPS")

	ips, err = parseIPStrings(ipStrings)
	if err != nil {
		return nil, fmt.Errorf("environment variable ALLOWED_IPS: %w", err)
	}

	return ips, nil
}

// getBlockedIPs obtains a list of IP addresses to block from
// the comma separated list for the environment variable BLOCK_IPS.
func getBlockedIPs() (ips []netaddr.IP, err error) {
	values := envToCSV("BLOCK_IPS")

	ips, err = parseIPStrings(values)
	if err != nil {
		return nil, fmt.Errorf("environment variable BLOCK_IPS: %w", err)
	}

	return ips, nil
}

// getAllowedIPPrefixes obtains a list of IP Prefixes to unblock from block lists
// from the comma separated list for the environment variable ALLOWED_CIDRS.
func getAllowedIPPrefixes() (ipPrefixes []netaddr.IPPrefix, err error) {
	ipPrefixStrings := envToCSV("ALLOWED_CIDRS")

	ipPrefixes, err = parseIPPrefixStrings(ipPrefixStrings)
	if err != nil {
		return nil, fmt.Errorf("environment variable ALLOWED_CIDRS: %w", err)
	}

	return ipPrefixes, nil
}

// getBlockedIPPrefixes obtains a list of IP networks (CIDR notation) to block from
// the comma separated list for the environment variable BLOCK_CIDRS.
func getBlockedIPPrefixes() (ipPrefixes []netaddr.IPPrefix, err error) {
	values := envToCSV("BLOCK_CIDRS")

	ipPrefixes, err = parseIPPrefixStrings(values)
	if err != nil {
		return nil, fmt.Errorf("environment variable BLOCK_CIDRS: %w", err)
	}

	return ipPrefixes, nil
}
