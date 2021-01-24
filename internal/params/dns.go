package params

import (
	"fmt"
	"net"

	"github.com/qdm12/dns/pkg/unbound"
	libparams "github.com/qdm12/golibs/params"
)

// GetProviders obtains the DNS over TLS providers to use
// from the environment variable PROVIDERS and PROVIDER for retro-compatibility.
func (r *reader) GetProviders() (providers []string, err error) {
	words, err := r.envParams.CSV("PROVIDERS", libparams.Default("cloudflare"),
		libparams.RetroKeys([]string{"PROVIDER"}, r.onRetroActive))
	if err != nil {
		return nil, err
	}

	for _, word := range words {
		provider := word
		if _, ok := unbound.GetProviderData(provider); !ok {
			return nil, fmt.Errorf("DNS provider %q is not valid", provider)
		}
		providers = append(providers, provider)
	}
	return providers, nil
}

// GetPrivateAddresses obtains if Unbound caching should be enable or not
// from the environment variable PRIVATE_ADDRESS.
func (r *reader) GetPrivateAddresses() (privateAddresses []string, err error) {
	privateAddresses, err = r.envParams.CSV("PRIVATE_ADDRESS")
	if err != nil {
		return nil, err
	}
	for _, address := range privateAddresses {
		ip := net.ParseIP(address)
		_, _, err := net.ParseCIDR(address)
		if ip == nil && err != nil {
			return nil, fmt.Errorf("private address %q is not a valid IP or CIDR range", address)
		}
	}
	return privateAddresses, nil
}
