package config

import (
	"strings"

	"github.com/qdm12/dns/pkg/provider"
	"github.com/qdm12/golibs/params"
)

// getProviders obtains the DNS over TLS providers to use
// from the environment variable PROVIDERS and PROVIDER for retro-compatibility.
func getProviders(reader *reader) (providers []provider.Provider, err error) {
	words, err := reader.env.CSV("PROVIDERS", params.Default("cloudflare"),
		params.RetroKeys([]string{"PROVIDER"}, reader.onRetroActive))
	if err != nil {
		return nil, err
	}

	for _, word := range words {
		// Retro compatibility
		word = strings.ReplaceAll(word, ".", " ")
		switch strings.ToLower(word) {
		case "cleanbrowsing":
			word = "cleanbrowsing security"
		case "cira":
			word = "cira private"
		}

		provider, err := provider.Parse(word)
		if err != nil {
			return nil, err
		}

		providers = append(providers, provider)
	}
	return providers, nil
}
