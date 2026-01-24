package dot

import (
	"context"
	"crypto/tls"
	"fmt"
	"hash/maphash"
	"maps"
	"math/rand/v2"
	"net"
	"net/netip"
	"slices"
)

type Dialer struct {
	addressToServerName map[string]string
	addrPorts           []netip.AddrPort
	ipv6                bool
	netDialer           *net.Dialer
	disablePooling      bool
	metrics             Metrics
	randGen             *rand.Rand
}

func New(settings Settings) (dial *Dialer, err error) {
	settings.SetDefaults()
	err = settings.Validate()
	if err != nil {
		return nil, fmt.Errorf("validating settings: %w", err)
	}

	addressToServerName := make(map[string]string, len(settings.UpstreamResolvers))
	addrPorts := make([]netip.AddrPort, 0, len(settings.UpstreamResolvers))
	for _, upstreamResolver := range settings.UpstreamResolvers {
		for _, addrPort := range upstreamResolver.DoT.IPv4 {
			addressToServerName[addrPort.String()] = upstreamResolver.DoT.Name
			addrPorts = append(addrPorts, addrPort)
		}
		if settings.IPVersion == "ipv6" {
			for _, addrPort := range upstreamResolver.DoT.IPv6 {
				addressToServerName[addrPort.String()] = upstreamResolver.DoT.Name
				addrPorts = append(addrPorts, addrPort)
			}
		}
	}

	return &Dialer{
		addressToServerName: addressToServerName,
		addrPorts:           addrPorts,
		ipv6:                settings.IPVersion == "ipv6",
		netDialer: &net.Dialer{
			Timeout: settings.Timeout,
		},
		disablePooling: settings.DisablePooling,
		metrics:        settings.Metrics,
		randGen:        rand.New(&mapHashSource{}), //nolint:gosec
	}, nil
}

func (d *Dialer) String() string {
	return "tls"
}

// ReusableConnsSupported returns true to indicate that connections created
// by this dialer can be reused in a connection pool. Returns false if
// DisablePooling was set in settings, which forces a new connection for
// each DNS query.
func (d *Dialer) ReusableConnsSupported() bool {
	return !d.disablePooling
}

func (d *Dialer) Addresses() []string {
	addresses := slices.Collect(maps.Keys(d.addressToServerName))
	slices.Sort(addresses)
	return addresses
}

func (d *Dialer) Dial(ctx context.Context, _, address string) (
	conn net.Conn, err error,
) {
	serverName, serverAddress := d.pickNameAddress(address)

	conn, err = d.netDialer.DialContext(ctx, "tcp", serverAddress)
	if err != nil {
		d.metrics.DoTDialInc(serverName, serverAddress, "dial error")
		return nil, fmt.Errorf("dialing tcp %s: %w", serverAddress, err)
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: serverName,
	}
	tlsConn := tls.Client(conn, tlsConfig)

	err = tlsConn.HandshakeContext(ctx)
	if err != nil {
		d.metrics.DoTDialInc(serverName, serverAddress, "handshake error")
		return nil, fmt.Errorf("running TLS handshake with %s (%s): %w",
			serverAddress, serverName, err)
	}
	d.metrics.DoTDialInc(serverName, serverAddress, "success")
	return tlsConn, nil
}

func (d *Dialer) pickNameAddress(dialAddr string) (name, address string) {
	serverName, ok := d.addressToServerName[dialAddr]
	if ok {
		return serverName, dialAddr
	}
	// when used as the dialer for a Go resolver, the address is the DNS system IP address
	// so pick a server address at random.
	address = d.addrPorts[d.randGen.IntN(len(d.addrPorts))].String()
	name = d.addressToServerName[address]
	return name, address
}

type mapHashSource struct{}

func (s *mapHashSource) Uint64() uint64 {
	return new(maphash.Hash).Sum64()
}
