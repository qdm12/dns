package env

import (
	"fmt"

	"github.com/qdm12/dns/v2/internal/config/settings"
)

func readDoT() (settings settings.DoT, err error) {
	settings.DoTProviders = envToCSV("DOT_RESOLVERS")
	settings.DNSProviders = envToCSV("DNS_FALLBACK_PLAINTEXT_RESOLVERS")
	settings.Timeout, err = envToDuration("DOT_TIMEOUT")
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_TIMEOUT: %w", err)
	}

	settings.IPv6, err = envToBoolPtr("DOT_CONNECT_IPV6")
	if err != nil {
		return settings, fmt.Errorf("environment variable DOT_CONNECT_IPV6: %w", err)
	}

	return settings, nil
}
