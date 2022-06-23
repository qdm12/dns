package env

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

func readDoH() (settings settings.DoH, err error) {
	settings.DoHProviders = envToCSV("DOH_RESOLVERS")
	settings.Timeout, err = envToDuration("DOH_TIMEOUT")
	if err != nil {
		return settings, fmt.Errorf("environment variable DOH_TIMEOUT: %w", err)
	}

	settings.Self.DoTProviders = envToCSV("DOT_RESOLVERS")
	settings.Self.DNSProviders = envToCSV("DNS_FALLBACK_PLAINTEXT_RESOLVERS")
	settings.Self.IPv6, err = envToBoolPtr("DOT_CONNECT_IPV6")
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_CONNECT_IPV6: %w", err)
	}

	settings.Self.Timeout, err = envToDuration("DOT_TIMEOUT")
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_TIMEOUT: %w", err)
	}

	return settings, nil
}
