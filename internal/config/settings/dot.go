package settings

import (
	"errors"
	"fmt"
	"time"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gotree"
)

type DoT struct {
	DoTProviders []string
	DNSProviders []string
	Timeout      time.Duration
	IPv6         *bool
}

func (d *DoT) setDefaults() {
	d.DoTProviders = gosettings.DefaultSlice(d.DoTProviders, []string{
		provider.Cloudflare().Name,
		provider.Google().Name,
	})

	d.DNSProviders = gosettings.DefaultSlice(d.DNSProviders, []string{
		provider.Cloudflare().Name,
		provider.Google().Name,
	})

	d.Timeout = gosettings.DefaultNumber(d.Timeout, time.Second)

	const defaultIPv6 = false // some systems do not support IPv6
	d.IPv6 = gosettings.DefaultPointer(d.IPv6, defaultIPv6)
}

var (
	ErrTimeoutTooSmall = errors.New("timeout is too small")
)

func (d *DoT) validate() (err error) {
	err = checkProviderNames(d.DoTProviders)
	if err != nil {
		return fmt.Errorf("DoT provider: %w", err)
	}

	err = checkProviderNames(d.DNSProviders)
	if err != nil {
		return fmt.Errorf("fallback DNS plaintext provider: %w", err)
	}

	const minTimeout = time.Millisecond
	if d.Timeout < minTimeout {
		return fmt.Errorf("%w: %s must be at least %s",
			ErrTimeoutTooSmall, d.Timeout, minTimeout)
	}

	return nil
}

func (d *DoT) String() string {
	return d.ToLinesNode().String()
}

func (d *DoT) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DNS over TLS:")

	node.Appendf("DNS over TLS providers: %s", andStrings(d.DoTProviders))

	if len(d.DNSProviders) > 0 {
		node.Appendf("Plaintext fallback DNS providers: %s", andStrings(d.DNSProviders))
	}

	node.Appendf("Request timeout: %s", d.Timeout)

	connectOver := "IPv4"
	if *d.IPv6 {
		connectOver = "IPv6"
	}
	node.Appendf("Connecting over: %s", connectOver)

	return node
}
