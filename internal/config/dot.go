package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gotree"
)

type DoT struct {
	DoTProviders []string
	Timeout      time.Duration
	IPv6         *bool
}

func (d *DoT) setDefaults() {
	d.DoTProviders = gosettings.DefaultSlice(d.DoTProviders, []string{
		provider.Cloudflare().Name,
		provider.Google().Name,
	})

	d.Timeout = gosettings.DefaultComparable(d.Timeout, time.Second)

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

	node.Appendf("Request timeout: %s", d.Timeout)

	connectOver := "IPv4"
	if *d.IPv6 {
		connectOver = "IPv6"
	}
	node.Appendf("Connecting over: %s", connectOver)

	return node
}

func (d *DoT) read(reader *reader.Reader) (err error) {
	d.DoTProviders = reader.CSV("DOT_RESOLVERS")
	d.Timeout, err = reader.Duration("DOT_TIMEOUT")
	if err != nil {
		return err
	}

	d.IPv6, err = reader.BoolPtr("DOT_CONNECT_IPV6")
	if err != nil {
		return err
	}

	return nil
}
