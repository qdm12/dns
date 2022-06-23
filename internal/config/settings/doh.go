package settings

import (
	"errors"
	"fmt"
	"time"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/gotree"
)

type DoH struct {
	DoHProviders []string
	Timeout      time.Duration
	Self         DoT
}

func (d *DoH) setDefaults() {
	if len(d.DoHProviders) == 0 {
		d.DoHProviders = []string{
			provider.Cloudflare().Name,
			provider.Google().Name,
		}
	}

	if d.Timeout == 0 {
		d.Timeout = time.Second
	}

	d.Self.setDefaults()
}

var (
	ErrDoHProviderNotValid = errors.New("DoH provider is not valid")
)

func (d *DoH) validate() (err error) {
	allProviders := provider.All()
	allProvidersSet := make(map[string]struct{}, len(allProviders))
	for _, provider := range allProviders {
		allProvidersSet[provider.Name] = struct{}{}
	}

	for _, provider := range d.DoHProviders {
		if _, ok := allProvidersSet[provider]; !ok {
			return fmt.Errorf("%w: %s", ErrDoHProviderNotValid, provider)
		}
	}

	const minTimeout = time.Millisecond
	if d.Timeout < minTimeout {
		return fmt.Errorf("%w: %s must be at least %s",
			ErrTimeoutTooSmall, d.Timeout, minTimeout)
	}

	err = d.Self.validate()
	if err != nil {
		return fmt.Errorf("self dns: %w", err)
	}

	return nil
}

func (d *DoH) String() string {
	return d.ToLinesNode().String()
}

func (d *DoH) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DNS over HTTPs:")

	node.Appendf("DNS over HTTPs providers: %s", andStrings(d.DoHProviders))

	node.Appendf("Request timeout: %s", d.Timeout)

	node.AppendNode(d.Self.ToLinesNode())

	return node
}
