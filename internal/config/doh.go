package config

import (
	"fmt"

	"github.com/qdm12/dns/pkg/doh"
	"github.com/qdm12/golibs/params"
)

func getDoHSettings(reader *reader) (settings doh.ResolverSettings, err error) {
	settings.DoHProviders, err = getDoHProviders(reader)
	if err != nil {
		return settings, err
	}

	settings.SelfDNS.DoTProviders, err = getDoTProviders(reader)
	if err != nil {
		return settings, err
	}

	settings.SelfDNS.DNSProviders, err = getDNSProviders(reader)
	if err != nil {
		return settings, err
	}

	settings.SelfDNS.IPv6, err = reader.env.OnOff("DOT_CONNECT_IPV6", params.Default("off"))
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_CONNECT_IPV6: %w", err)
	}

	settings.SelfDNS.Timeout, err = reader.env.Duration("DOT_TIMEOUT", params.Default("3s"))
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_TIMEOUT: %w", err)
	}

	settings.Timeout, err = reader.env.Duration("DOH_TIMEOUT", params.Default("3s"))
	if err != nil {
		return settings, fmt.Errorf("environment variable DOH_TIMEOUT: %w", err)
	}

	return settings, nil
}
