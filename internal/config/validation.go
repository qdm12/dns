package config

import (
	"github.com/qdm12/dns/v2/pkg/provider"
)

func checkUpstreamResolverNames(providerNames []string) (err error) {
	providers := provider.NewProviders()
	for _, providerName := range providerNames {
		_, err := providers.Get(providerName)
		if err != nil {
			return err
		}
	}

	return nil
}
