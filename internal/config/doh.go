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
	// UpstreamResolvers is a list of DNS over HTTPS upstream
	// resolvers to use.
	UpstreamResolvers []string
	// IPVersion defines the only IP version to use to connect to
	// upstream DNS over HTTPS servers. If left unset, it defaults
	// to "ipv4".
	IPVersion string
	// Timeout is the maximum duration to wait for a response from
	// upstream DNS over HTTPS servers. If left unset, it defaults
	// to 1 second.
	Timeout time.Duration
}

func (d *DoH) setDefaults() {
	d.UpstreamResolvers = gosettings.DefaultSlice(d.UpstreamResolvers, []string{
		provider.Cloudflare().Name,
		provider.Google().Name,
	})
	d.IPVersion = gosettings.DefaultComparable(d.IPVersion, "ipv4")
	d.Timeout = gosettings.DefaultComparable(d.Timeout, time.Second)
}

func (d *DoH) validate() (err error) {
	err = checkUpstreamResolverNames(d.UpstreamResolvers)
	if err != nil {
		return fmt.Errorf("upstream resolvers: %w", err)
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

	node.Appendf("Upstream resolvers: %s", andStrings(d.UpstreamResolvers))
	node.Appendf("Connecting over %s", d.IPVersion)
	node.Appendf("Query timeout: %s", d.Timeout)

	return node
}

func (d *DoH) read(reader *reader.Reader) (err error) {
	d.UpstreamResolvers = reader.CSV("DOH_RESOLVERS")
	d.IPVersion = reader.String("DOH_IP_VERSION")

	d.Timeout, err = reader.Duration("DOH_TIMEOUT")
	if err != nil {
		return err
	}

	return nil
}
