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
func (p *paramsReader) GetProviders() (providers []models.Provider, err error) {
	// Retro-compatibility
	s, err := p.envParams.GetEnv("PROVIDER")
	if err != nil {
		return nil, err
	} else if len(s) != 0 {
		p.logger.Warn("You are using the old environment variable PROVIDER, please consider changing it to PROVIDERS")
	} else {
		s, err = p.envParams.GetEnv("PROVIDERS", libparams.Default("cloudflare"))
		if err != nil {
			return nil, err
		}
	}
	for _, word := range strings.Split(s, ",") {
		provider := models.Provider(word)
		switch provider {
		case constants.Cloudflare, constants.Google, constants.Quad9, constants.Quadrant, constants.CleanBrowsing, constants.SecureDNS, constants.LibreDNS:
			providers = append(providers, provider)
		default:
			return nil, fmt.Errorf("DNS over TLS provider %q is not valid", provider)
		}
	}
	return providers, nil
}

// GetPrivateAddresses obtains if Unbound caching should be enable or not
// from the environment variable PRIVATE_ADDRESS
func (p *paramsReader) GetPrivateAddresses() (privateAddresses []string) {
	s, _ := p.envParams.GetEnv("PRIVATE_ADDRESS")
	for _, s := range strings.Split(s, ",") {
		privateAddresses = append(privateAddresses, s)
	}
	return privateAddresses
}
