package provider

import (
	"errors"
	"fmt"
	"strings"
)

type Providers struct {
	providers []Provider
}

func NewProviders() *Providers {
	return &Providers{
		providers: []Provider{
			CiraFamily(),
			CiraPrivate(),
			CiraProtected(),
			CleanBrowsingAdult(),
			CleanBrowsingFamily(),
			CleanBrowsingSecurity(),
			Cloudflare(),
			CloudflareFamily(),
			CloudflareSecurity(),
			Google(),
			LibreDNS(),
			OpenDNS(),
			Quad9(),
			Quad9Secured(),
			Quad9Unsecured(),
			Quadrant(),
		},
	}
}

func (p *Providers) List() (providers []Provider) {
	providers = make([]Provider, len(p.providers))
	copy(providers, p.providers)
	return providers
}

var ErrParseProviderNameUnknown = errors.New("provider does not match any known providers")

func (p *Providers) Get(name string) (provider Provider, err error) {
	for _, provider = range p.providers {
		if strings.EqualFold(name, provider.Name) {
			return provider, nil
		}
	}

	return Provider{}, fmt.Errorf("%w: %s", ErrParseProviderNameUnknown, name)
}
