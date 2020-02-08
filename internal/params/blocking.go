package params

import (
	"fmt"
	"strings"

	libparams "github.com/qdm12/golibs/params"
)

// GetMaliciousBlocking obtains if malicious hostnames/IPs should be blocked
// from being resolved by Unbound, using the environment variable BLOCK_MALICIOUS
func (p *paramsReader) GetMaliciousBlocking() (blocking bool, err error) {
	return p.envParams.GetOnOff("BLOCK_MALICIOUS", libparams.Default("on"))
}

// GetSurveillanceBlocking obtains if surveillance hostnames/IPs should be blocked
// from being resolved by Unbound, using the environment variable BLOCK_SURVEILLANCE
// and BLOCK_NSA for retrocompatibility
func (p *paramsReader) GetSurveillanceBlocking() (blocking bool, err error) {
	// Retro-compatibility
	s, err := p.envParams.GetEnv("BLOCK_NSA")
	if err != nil {
		return false, err
	} else if len(s) != 0 {
		p.logger.Warn("You are using the old environment variable BLOCK_NSA, please consider changing it to BLOCK_SURVEILLANCE")
		return p.envParams.GetOnOff("BLOCK_NSA", libparams.Compulsory())
	}
	return p.envParams.GetOnOff("BLOCK_SURVEILLANCE", libparams.Default("off"))
}

// GetAdsBlocking obtains if ads hostnames/IPs should be blocked
// from being resolved by Unbound, using the environment variable BLOCK_ADS
func (p *paramsReader) GetAdsBlocking() (blocking bool, err error) {
	return p.envParams.GetOnOff("BLOCK_ADS", libparams.Default("off"))
}

// GetUnblockedHostnames obtains a list of hostnames to unblock from block lists
// from the comma separated list for the environment variable UNBLOCK
func (p *paramsReader) GetUnblockedHostnames() (hostnames []string, err error) {
	s, err := p.envParams.GetEnv("UNBLOCK")
	if err != nil {
		return nil, err
	}
	if len(s) == 0 {
		return nil, nil
	}
	hostnames = strings.Split(s, ",")
	for _, hostname := range hostnames {
		if !p.verifier.MatchHostname(hostname) {
			return nil, fmt.Errorf("hostname %q does not seem valid", hostname)
		}
	}
	return hostnames, nil
}
