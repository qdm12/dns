package env

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

func (r *Reader) readDoH() (settings settings.DoH, err error) {
	settings.DoHProviders = r.env.CSV("DOH_RESOLVERS")
	settings.Timeout, err = r.env.Duration("DOH_TIMEOUT")
	if err != nil {
		return settings, fmt.Errorf("environment variable DOH_TIMEOUT: %w", err)
	}

	settings.Self.DoTProviders = r.env.CSV("DOT_RESOLVERS")
	settings.Self.DNSProviders = r.env.CSV("DNS_FALLBACK_PLAINTEXT_RESOLVERS")
	settings.Self.IPv6, err = r.env.BoolPtr("DOT_CONNECT_IPV6")
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_CONNECT_IPV6: %w", err)
	}

	settings.Self.Timeout, err = r.env.Duration("DOT_TIMEOUT")
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_TIMEOUT: %w", err)
	}

	return settings, nil
}
