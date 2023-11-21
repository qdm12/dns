package env

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

func (r *Reader) readDoT() (settings settings.DoT, err error) {
	settings.DoTProviders = r.reader.CSV("DOT_RESOLVERS")
	settings.DNSProviders = r.reader.CSV("DNS_FALLBACK_PLAINTEXT_RESOLVERS")
	settings.Timeout, err = r.reader.Duration("DOT_TIMEOUT")
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_TIMEOUT: %w", err)
	}

	settings.IPv6, err = r.reader.BoolPtr("DOT_CONNECT_IPV6")
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_CONNECT_IPV6: %w", err)
	}

	return settings, nil
}
