package params

import (
	"fmt"
	"strings"

	"github.com/qdm12/cloudflare-dns-server/internal/constants"
	"github.com/qdm12/cloudflare-dns-server/internal/models"
	libparams "github.com/qdm12/golibs/params"
)

// GetProviders obtains the DNS over TLS providers to use
// from the environment variable PROVIDERS and PROVIDER for retro-compatibility
func (r *reader) GetProviders() (providers []models.Provider, err error) {
	// Retro-compatibility
	s, err := r.envParams.GetEnv("PROVIDER")
	switch {
	case err != nil:
		return nil, err
	case len(s) != 0:
		r.logger.Warn("You are using the old environment variable PROVIDER, please consider changing it to PROVIDERS")
	default:
		s, err = r.envParams.GetEnv("PROVIDERS", libparams.Default("cloudflare"))
		if err != nil {
			return nil, err
		}
	}
	for _, word := range strings.Split(s, ",") {
		provider := models.Provider(word)
		if _, ok := constants.ProviderMapping()[provider]; !ok {
			return nil, fmt.Errorf("DNS provider %q is not valid", provider)
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

// GetPrivateAddresses obtains if Unbound caching should be enable or not
// from the environment variable PRIVATE_ADDRESS
func (r *reader) GetPrivateAddresses() (privateAddresses []string) {
	s, _ := r.envParams.GetEnv("PRIVATE_ADDRESS")
	privateAddresses = append(privateAddresses, strings.Split(s, ",")...)
	return privateAddresses
}
