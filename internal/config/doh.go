package config

import (
	"fmt"
	"time"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gosettings/validate"
	"github.com/qdm12/gotree"
)

type DoH struct {
	DoHProviders []string
	IPVersion    string
	Timeout      time.Duration
}

func (d *DoH) setDefaults() {
	d.DoHProviders = gosettings.DefaultSlice(d.DoHProviders, []string{
		provider.Cloudflare().Name,
		provider.Google().Name,
	})
	d.IPVersion = gosettings.DefaultComparable(d.IPVersion, "ipv4")
	d.Timeout = gosettings.DefaultComparable(d.Timeout, time.Second)
}

func (d *DoH) validate() (err error) {
	err = checkProviderNames(d.DoHProviders)
	if err != nil {
		return fmt.Errorf("DoH provider: %w", err)
	}

	err = validate.IsOneOf(d.IPVersion, "ipv4", "ipv6")
	if err != nil {
		return fmt.Errorf("IP version: %w", err)
	}

	const minTimeout = time.Millisecond
	if d.Timeout < minTimeout {
		return fmt.Errorf("%w: %s must be at least %s",
			ErrTimeoutTooSmall, d.Timeout, minTimeout)
	}

	return nil
}

func (d *DoH) String() string {
	return d.ToLinesNode().String()
}

func (d *DoH) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("DNS over HTTPs:")

	node.Appendf("DNS over HTTPs providers: %s", andStrings(d.DoHProviders))
	node.Appendf("Connecting over %s", d.IPVersion)
	node.Appendf("Query timeout: %s", d.Timeout)

	return node
}

func (d *DoH) read(reader *reader.Reader) (err error) {
	d.DoHProviders = reader.CSV("DOH_RESOLVERS")
	d.IPVersion = reader.String("DOH_IP_VERSION")

	d.Timeout, err = reader.Duration("DOH_TIMEOUT")
	if err != nil {
		return err
	}

	return nil
}
