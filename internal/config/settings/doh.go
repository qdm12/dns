package settings

import (
	"fmt"
	"time"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type DoH struct {
	DoHProviders []string
	Timeout      time.Duration
	Self         DoT
}

func (d *DoH) setDefaults() {
	d.DoHProviders = gosettings.DefaultSlice(d.DoHProviders, []string{
		provider.Cloudflare().Name,
		provider.Google().Name,
	})
	d.Timeout = gosettings.DefaultNumber(d.Timeout, time.Second)
	d.Self.setDefaults()
}

func (d *DoH) validate() (err error) {
	err = checkProviderNames(d.DoHProviders)
	if err != nil {
		return fmt.Errorf("DoH provider: %w", err)
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
