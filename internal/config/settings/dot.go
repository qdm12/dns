package settings

import (
	"errors"
	"fmt"
	"time"

	"github.com/qdm12/dns/v2/internal/config/defaults"
	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/gotree"
)

type DoT struct {
	DoTProviders []string
	DNSProviders []string
	Timeout      time.Duration
	IPv6         *bool
}

func (d *DoT) setDefaults() {
	if len(d.DoTProviders) == 0 {
		d.DoTProviders = []string{
			provider.Cloudflare().Name,
			provider.Google().Name,
		}
	}

	if len(d.DNSProviders) == 0 {
		d.DNSProviders = []string{
			provider.Cloudflare().Name,
			provider.Google().Name,
		}
	}

	if d.Timeout == 0 {
		d.Timeout = time.Second
	}

	if d.IPv6 == nil {
		const defaultIPv6 = false // some systems do not support IPv6
		d.IPv6 = defaults.BoolPtr(d.IPv6, defaultIPv6)
	}
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
