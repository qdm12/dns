package config

import (
	"fmt"

	"github.com/qdm12/dns/v2/pkg/dot"
	"github.com/qdm12/golibs/params"
)

func getDoTSettings(reader *Reader) (settings dot.ResolverSettings, err error) {
	settings.DoTProviders, err = getDoTProviders(reader)
	if err != nil {
		return settings, err
	}

	settings.DNSProviders, err = getDNSProviders(reader)
	if err != nil {
		return settings, err
	}

	settings.Timeout, err = reader.env.Duration("DOT_TIMEOUT", params.Default("3s"))
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_TIMEOUT: %w", err)
	}

	ipv6, err := reader.env.OnOff("DOT_CONNECT_IPV6", params.Default("off"))
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_CONNECT_IPV6: %w", err)
	}
	settings.IPv6 = &ipv6

	return settings, nil
}
