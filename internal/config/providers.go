package config

import (
	"fmt"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/golibs/params"
)

// getDoTProviders obtains the DoT resolver providers to use.
func getDoTProviders(reader *Reader) (providers []string, err error) {
	return getProviders(reader, "DOT_RESOLVERS", "cloudflare,google")
}

// getDoHProviders obtains the DoH resolver providers to use.
func getDoHProviders(reader *Reader) (providers []string, err error) {
	return getProviders(reader, "DOH_RESOLVERS", "cloudflare,google")
}

// getDNSProviders obtains the plaintext fallback DNS resolver providers to use.
func getDNSProviders(reader *Reader) (providers []string, err error) {
	return getProviders(reader, "DNS_FALLBACK_PLAINTEXT_RESOLVERS", "")
}

// getProviders obtains the DNS resolver providers to use from the environment
// variable specified by key.
func getProviders(reader *Reader, key, defaultValue string) (providers []string, err error) {
	providers, err = reader.env.CSV(key, params.Default(defaultValue))
	if err != nil {
		return nil, fmt.Errorf("environment variable %s: %w", key, err)
	}

	for _, s := range providers {
		_, err := provider.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("environment variable %s: %w", key, err)
		}
	}
	return providers, nil
}
