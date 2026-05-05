package config

import (
	"fmt"

	"github.com/qdm12/dns/v2/pkg/provider"
)

func checkUpstreamResolverNames(providerNames []string, upstreamType string, ipv6 bool) (err error) {
	providers := provider.NewProviders()
	for _, providerName := range providerNames {
		provider, err := providers.Get(providerName)
		if err != nil {
			return err
		}
		err = provider.ValidateForType(upstreamType, ipv6)
		if err != nil {
			return fmt.Errorf("validating %s for %s: %w", providerName, upstreamType, err)
		}
	}

	return nil
}
