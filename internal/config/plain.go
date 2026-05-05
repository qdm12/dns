package config

import (
	"fmt"
	"time"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/qdm12/gosettings"
	"github.com/qdm12/gosettings/reader"
	"github.com/qdm12/gotree"
)

type Plain struct {
	// UpstreamResolvers is a list of DNS upstream resolvers to use.
	UpstreamResolvers []string
	// Timeout is the maximum duration to wait for a response from
	// upstream plaintext DNS servers. If left unset, it defaults to
	// 1 second.
	Timeout time.Duration
}

func (p *Plain) setDefaults() {
	p.UpstreamResolvers = gosettings.DefaultSlice(p.UpstreamResolvers, []string{
		provider.Cloudflare().Name,
		provider.Google().Name,
	})

	const defaultTimeout = 3 * time.Second
	p.Timeout = gosettings.DefaultComparable(p.Timeout, defaultTimeout)
}

func (p *Plain) validate(ipv6 bool) (err error) {
	err = checkUpstreamResolverNames(p.UpstreamResolvers, "plain", ipv6)
	if err != nil {
		return fmt.Errorf("upstream resolvers: %w", err)
	}

	const minTimeout = time.Millisecond
	if p.Timeout < minTimeout {
		return fmt.Errorf("%w: %s must be at least %s",
			ErrTimeoutTooSmall, p.Timeout, minTimeout)
	}

	return nil
}

func (p *Plain) String() string {
	return p.ToLinesNode().String()
}

func (p *Plain) ToLinesNode() (node *gotree.Node) {
	node = gotree.New("Plaintext:")

	node.Appendf("Upstream resolvers: %s", andStrings(p.UpstreamResolvers))
	node.Appendf("Request timeout: %s", p.Timeout)

	return node
}

func (p *Plain) read(reader *reader.Reader) (err error) {
	p.UpstreamResolvers = reader.CSV("PLAIN_RESOLVERS")
	p.Timeout, err = reader.Duration("PLAIN_TIMEOUT")
	if err != nil {
		return err
	}

	return nil
}
