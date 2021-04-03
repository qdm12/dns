package config

import (
	"fmt"
	"net"
	"strings"

	"github.com/qdm12/dns/pkg/provider"
	libparams "github.com/qdm12/golibs/params"
)

// GetProviders obtains the DNS over TLS providers to use
// from the environment variable PROVIDERS and PROVIDER for retro-compatibility.
func (r *reader) GetProviders() (providers []provider.Provider, err error) {
	words, err := r.envParams.CSV("PROVIDERS", libparams.Default("cloudflare"),
		libparams.RetroKeys([]string{"PROVIDER"}, r.onRetroActive))
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
