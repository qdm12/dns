package config

import (
	"fmt"

	"github.com/qdm12/dns/pkg/blacklist"
	"github.com/qdm12/golibs/params"
	"inet.af/netaddr"
)

func getBlacklistSettings(reader *reader) (settings blacklist.BuilderSettings, err error) {
	settings.BlockMalicious, err = reader.env.OnOff("BLOCK_MALICIOUS", params.Default("on"))
	if err != nil {
		return settings, err
	}
	settings.BlockSurveillance, err = reader.env.OnOff("BLOCK_SURVEILLANCE", params.Default("off"),
		params.RetroKeys([]string{"BLOCK_NSA"}, reader.onRetroActive))
	if err != nil {
		return settings, err
	}
	settings.BlockAds, err = reader.env.OnOff("BLOCK_ADS", params.Default("off"))
	if err != nil {
		return settings, err
	}
	settings.AllowedHosts, err = getAllowedHostnames(reader)
	if err != nil {
		return settings, err
	}
	settings.AddBlockedHosts, err = getBlockedHostnames(reader)
	if err != nil {
		return settings, err
	}
	settings.AddBlockedIPs, settings.AddBlockedIPPrefixes, err = getBlockedIPs(reader)
	if err != nil {
		return settings, err
	}
	privateIPs, privateIPPrefixes, err := getPrivateAddresses(reader)
	if err != nil {
		return settings, err
	}
	settings.AddBlockedIPs = append(settings.AddBlockedIPs, privateIPs...)
	settings.AddBlockedIPPrefixes = append(settings.AddBlockedIPPrefixes, privateIPPrefixes...)
	return settings, nil
}

// getAllowedHostnames obtains a list of hostnames to unblock from block lists
// from the comma separated list for the environment variable UNBLOCK.
func getAllowedHostnames(reader *reader) (hostnames []string, err error) {
	hostnames, err = reader.env.CSV("UNBLOCK")
	if err != nil {
		return nil, err
	}
	for _, hostname := range hostnames {
		if !reader.verifier.MatchHostname(hostname) {
			return nil, fmt.Errorf("unblocked hostname %q does not seem valid", hostname)
		}
	}
	return hostnames, nil
}

// getBlockedHostnames obtains a list of hostnames to block from the comma
// separated list for the environment variable BLOCK_HOSTNAMES.
func getBlockedHostnames(reader *reader) (hostnames []string, err error) {
	hostnames, err = reader.env.CSV("BLOCK_HOSTNAMES")
	if err != nil {
		return nil, err
	}
	for _, hostname := range hostnames {
		if !reader.verifier.MatchHostname(hostname) {
			return nil, fmt.Errorf("blocked hostname %q does not seem valid", hostname)
		}
	}
	return hostnames, nil
}

// getBlockedIPs obtains a list of IP addresses and IP networks to block from
// the comma separated list for the environment variable BLOCK_IPS.
func getBlockedIPs(reader *reader) (ips []netaddr.IP,
	ipPrefixes []netaddr.IPPrefix, err error) {
	values, err := reader.env.CSV("BLOCK_IPS")
	if err != nil {
		return nil, nil, err
	}
	ips, ipPrefixes, err = convertStringsToIPs(values)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid blocked IP string: %s", err)
	}
	return ips, ipPrefixes, nil
}

func convertStringsToIPs(values []string) (ips []netaddr.IP,
	ipPrefixes []netaddr.IPPrefix, err error) {
	ips = make([]netaddr.IP, 0, len(values))
	ipPrefixes = make([]netaddr.IPPrefix, 0, len(values))
	for _, value := range values {
		ip, err := netaddr.ParseIP(value)
		if err == nil {
			ips = append(ips, ip)
			continue
		}
		ipPrefix, err := netaddr.ParseIPPrefix(value)
		if err == nil {
			ipPrefixes = append(ipPrefixes, ipPrefix)
			continue
		}
		return nil, nil, fmt.Errorf("%s", value)
	}
	return ips, ipPrefixes, nil
}
