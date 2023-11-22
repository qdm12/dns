package config

import (
	"fmt"
	"time"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
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
	d.Timeout = gosettings.DefaultComparable(d.Timeout, time.Second)
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

func (d *DoH) read(reader *reader.Reader) (err error) {
	d.DoHProviders = reader.CSV("DOH_RESOLVERS")
	d.Timeout, err = reader.Duration("DOH_TIMEOUT")
	if err != nil {
		return err
	}

	d.Self.DoTProviders = reader.CSV("DOT_RESOLVERS")
	d.Self.DNSProviders = reader.CSV("DNS_FALLBACK_PLAINTEXT_RESOLVERS")
	d.Self.IPv6, err = reader.BoolPtr("DOT_CONNECT_IPV6")
	if err != nil {
		return err
	}

	d.Self.Timeout, err = reader.Duration("DOT_TIMEOUT")
	if err != nil {
		return err
	}

	return nil
}
