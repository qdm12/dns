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

func (r *reader) GetPrivateAddresses() (privateIPs []net.IP, privateIPNets []*net.IPNet, err error) {
	values, err := r.envParams.CSV("PRIVATE_ADDRESS")
	if err != nil {
		return nil, nil, err
	}
	privateIPs, privateIPNets, err = convertStringsToIPs(values)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid private IP string: %s", err)
	}
	return privateIPs, privateIPNets, nil
}
