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
	// UpstreamResolvers is a list of DNS over TLS upstream
	// resolvers to use.
	UpstreamResolvers []string
	// Timeout is the maximum duration to wait for a response from
	// upstream DNS over TLS servers. If left unset, it defaults to
	// 1 second.
	Timeout   time.Duration
	IPVersion string
}

func (d *DoT) setDefaults() {
	d.UpstreamResolvers = gosettings.DefaultSlice(d.UpstreamResolvers, []string{
		provider.Cloudflare().Name,
		provider.Google().Name,
	})

	d.Timeout = gosettings.DefaultComparable(d.Timeout, time.Second)
	d.IPVersion = gosettings.DefaultComparable(d.IPVersion, "ipv4")
}

var (
	ErrTimeoutTooSmall = errors.New("timeout is too small")
)

func (d *DoT) validate() (err error) {
	err = checkUpstreamResolverNames(d.UpstreamResolvers)
	if err != nil {
		return fmt.Errorf("upstream resolvers: %w", err)
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

	node.Appendf("Upstream resolvers: %s", andStrings(d.UpstreamResolvers))
	node.Appendf("Request timeout: %s", d.Timeout)
	node.Appendf("Connecting over: %s", d.IPVersion)

	return node
}

func (d *DoT) read(reader *reader.Reader) (err error) {
	d.UpstreamResolvers = reader.CSV("DOT_RESOLVERS")
	d.Timeout, err = reader.Duration("DOT_TIMEOUT")
	if err != nil {
		return err
	}

	d.IPVersion = reader.String("DOT_IP_VERSION")
	return nil
}
