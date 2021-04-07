package config

import (
	"fmt"

	"github.com/qdm12/dns/pkg/provider"
	"github.com/qdm12/golibs/params"
)

// getDoTProviders obtains the DoT resolver providers to use.
func getDoTProviders(reader *reader) (providers []provider.Provider, err error) {
	return getProviders(reader, "DOT_PROVIDERS", "cloudflare,google")
}

// getDoHProviders obtains the DoH resolver providers to use.
func getDoHProviders(reader *reader) (providers []provider.Provider, err error) {
	return getProviders(reader, "DOH_PROVIDERS", "cloudflare,google")
}

// getDNSProviders obtains the plaintext fallback DNS resolver providers to use.
func getDNSProviders(reader *reader) (providers []provider.Provider, err error) {
	return getProviders(reader, "DNS_PLAINTEXT_PROVIDERS", "cloudflare")
}

// getProviders obtains the DNS resolver providers to use from the environment
// variable specified by key.
func getProviders(reader *reader, key, defaultValue string) (providers []provider.Provider, err error) {
	words, err := reader.env.CSV(key, params.Default(defaultValue))
	if err != nil {
		return nil, fmt.Errorf("environment variable PROVIDERS: %w", err)
	}

	providers = make([]provider.Provider, 0, len(words))
	for _, word := range words {
		provider, err := provider.Parse(word)
		if err != nil {
			return nil, fmt.Errorf("environment variable PROVIDERS: %w", err)
		}

		providers = append(providers, provider)
	}
	return providers, nil
}
