package plain

import (
	"context"
	"fmt"
	"hash/maphash"
	"maps"
	"math/rand/v2"
	"net"
	"net/netip"
	"slices"
)

type Dialer struct {
	addrStrings map[string]struct{}
	addrPorts   []netip.AddrPort
	ipv6        bool
	netDialer   *net.Dialer
	metrics     Metrics
	randGen     *rand.Rand
}

func New(settings Settings) (dial *Dialer, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	addrStrings := make(map[string]struct{}, len(settings.UpstreamResolvers))
	addrPorts := make([]netip.AddrPort, 0, len(settings.UpstreamResolvers))
	for _, upstreamResolver := range settings.UpstreamResolvers {
		for _, addrPort := range upstreamResolver.Plain.IPv4 {
			addrStrings[addrPort.String()] = struct{}{}
			addrPorts = append(addrPorts, addrPort)
		}
		if settings.IPVersion == "ipv6" {
			for _, addrPort := range upstreamResolver.Plain.IPv6 {
				addrStrings[addrPort.String()] = struct{}{}
				addrPorts = append(addrPorts, addrPort)
			}
		}
	}

	return &Dialer{
		addrStrings: addrStrings,
		addrPorts:   addrPorts,
		ipv6:        settings.IPVersion == "ipv6",
		netDialer: &net.Dialer{
			Timeout: settings.Timeout,
		},
		metrics: settings.Metrics,
		randGen: rand.New(&mapHashSource{}), //nolint:gosec
	}, nil
}

func (d *Dialer) String() string {
	return "plaintext"
}

// ReusableConnsSupported returns false to indicate that connections created
// by this dialer cannot be reused.
func (d *Dialer) ReusableConnsSupported() bool {
	return false
}

func (d *Dialer) Addresses() []string {
	addresses := slices.Collect(maps.Keys(d.addrStrings))
	slices.Sort(addresses)
	return addresses
}

func (d *Dialer) Dial(ctx context.Context, network, address string) (
	conn net.Conn, err error,
) {
	serverAddress := d.pickAddress(address)

	udpConn, err := d.netDialer.DialContext(ctx, network, serverAddress)
	if err != nil {
		d.metrics.PlainDialInc(serverAddress, "error")
		return nil, err
	}

	d.metrics.PlainDialInc(serverAddress, "success")
	return udpConn, nil
}

func (d *Dialer) pickAddress(dialAddr string) string {
	_, ok := d.addrStrings[dialAddr]
	if ok {
		return dialAddr
	}
	// when used as the dialer for a Go resolver, the address is the DNS system IP address
	// so pick a server address at random.
	return d.addrPorts[d.randGen.IntN(len(d.addrPorts))].String()
}

type mapHashSource struct{}

func (s *mapHashSource) Uint64() uint64 {
	return new(maphash.Hash).Sum64()
}
