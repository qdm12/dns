package config

import (
	"fmt"
	"net"

	libparams "github.com/qdm12/golibs/params"
)

// GetMaliciousBlocking obtains if malicious hostnames/IPs should be blocked
// from being resolved by Unbound, using the environment variable BLOCK_MALICIOUS.
func (r *reader) GetMaliciousBlocking() (blocking bool, err error) {
	return r.envParams.OnOff("BLOCK_MALICIOUS", libparams.Default("on"))
}

// GetSurveillanceBlocking obtains if surveillance hostnames/IPs should be blocked
// from being resolved by Unbound, using the environment variable BLOCK_SURVEILLANCE
// and BLOCK_NSA for retrocompatibility.
func (r *reader) GetSurveillanceBlocking() (blocking bool, err error) {
	return r.envParams.OnOff("BLOCK_SURVEILLANCE", libparams.Default("off"),
		libparams.RetroKeys([]string{"BLOCK_NSA"}, r.onRetroActive))
}

// GetAdsBlocking obtains if ads hostnames/IPs should be blocked
// from being resolved by Unbound, using the environment variable BLOCK_ADS.
func (r *reader) GetAdsBlocking() (blocking bool, err error) {
	return r.envParams.OnOff("BLOCK_ADS", libparams.Default("off"))
}

// GetUnblockedHostnames obtains a list of hostnames to unblock from block lists
// from the comma separated list for the environment variable UNBLOCK.
func (r *reader) GetUnblockedHostnames() (hostnames []string, err error) {
	hostnames, err = r.envParams.CSV("UNBLOCK")
	if err != nil {
		return nil, err
	}
	for _, hostname := range hostnames {
		if !r.verifier.MatchHostname(hostname) {
			return nil, fmt.Errorf("unblocked hostname %q does not seem valid", hostname)
		}
	}
	return hostnames, nil
}

// GetBlockedHostnames obtains a list of hostnames to block from the comma
// separated list for the environment variable BLOCK_HOSTNAMES.
func (r *reader) GetBlockedHostnames() (hostnames []string, err error) {
	hostnames, err = r.envParams.CSV("BLOCK_HOSTNAMES")
	if err != nil {
		return nil, err
	}
	for _, hostname := range hostnames {
		if !r.verifier.MatchHostname(hostname) {
			return nil, fmt.Errorf("blocked hostname %q does not seem valid", hostname)
		}
	}
	return hostnames, nil
}

// GetBlockedIPs obtains a list of IP addresses and IP networks to block from
// the comma separated list for the environment variable BLOCK_IPS.
func (r *reader) GetBlockedIPs() (ips []net.IP, ipNets []*net.IPNet, err error) {
	values, err := r.envParams.CSV("BLOCK_IPS")
	if err != nil {
		return nil, nil, err
	}
	ips, ipNets, err = convertStringsToIPs(values)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid blocked IP string: %s", err)
	}
	return ips, ipNets, nil
}

func convertStringsToIPs(values []string) (ips []net.IP, ipNets []*net.IPNet, err error) {
	ips = make([]net.IP, 0, len(values))
	ipNets = make([]*net.IPNet, 0, len(values))
	for _, value := range values {
		ip := net.ParseIP(value)
		if ip != nil {
			ips = append(ips, ip)
			continue
		}
		_, IPNet, err := net.ParseCIDR(value)
		if err == nil && IPNet != nil {
			ipNets = append(ipNets, IPNet)
			continue
		}
		return nil, nil, fmt.Errorf("%s", value)
	}
	return ips, ipNets, nil
}
