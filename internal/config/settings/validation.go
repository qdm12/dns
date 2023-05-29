package settings

import (
	"fmt"

	"github.com/qdm12/dns/v2/pkg/provider"
)

func checkProviderNames(providerNames []string) (err error) {
	for _, providerName := range providerNames {
		_, err := provider.Parse(providerName)
		if err != nil {
			return fmt.Errorf("parsing provider: %w", err)
		}
	}

	return nil
}
