package params

import (
	"fmt"
	"net"
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
func (r *reader) GetPrivateAddresses() (privateAddresses []string, err error) {
	s, err := r.envParams.GetEnv("PRIVATE_ADDRESS")
	if err != nil {
		return nil, err
	} else if len(s) == 0 {
		return nil, nil
	}
	privateAddresses = strings.Split(s, ",")
	for _, address := range privateAddresses {
		ip := net.ParseIP(address)
		_, _, err := net.ParseCIDR(address)
		if ip == nil && err != nil {
			return nil, fmt.Errorf("private address %q is not a valid IP or CIDR range", address)
		}
	}
	return privateAddresses, nil
}
