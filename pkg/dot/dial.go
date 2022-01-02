package dot

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/qdm12/dns/internal/picker"
	"github.com/qdm12/dns/internal/server"
	"github.com/qdm12/dns/pkg/dot/metrics"
	"github.com/qdm12/dns/pkg/log"
	"github.com/qdm12/dns/pkg/provider"
)

func newDoTDial(settings ResolverSettings) (
	dial server.Dial, err error) {
	warner := settings.Warner
	metrics := settings.Metrics

	dotServers, dnsServers, err := settingsToServers(settings)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{
		Timeout: settings.Timeout,
	}

	picker := picker.New()

	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		serverName, serverAddress := pickNameAddress(picker,
			dotServers, *settings.IPv6)

		conn, err := dialer.DialContext(ctx, "tcp", serverAddress)
		if err != nil {
			return onDialError(ctx, err, serverName, serverAddress, dialer,
				picker, *settings.IPv6, dnsServers, warner, metrics)
		}

		metrics.DoTDialInc(serverName, serverAddress, "success")

		tlsConf := &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: serverName,
		}
		// TODO handshake? See tls.DialWithDialer
		return tls.Client(conn, tlsConf), nil
	}, nil
}

func settingsToServers(settings ResolverSettings) (
	dotServers []provider.DoTServer,
	dnsServers []provider.DNSServer,
	err error) {
	dotServers = make([]provider.DoTServer, len(settings.DoTProviders))
	for i, s := range settings.DoTProviders {
		provider, err := provider.Parse(s)
		if err != nil {
			return nil, nil, err
		}
		dotServers[i] = provider.DoT()
	}

	dnsServers = make([]provider.DNSServer, len(settings.DNSProviders))
	for i, s := range settings.DNSProviders {
		provider, err := provider.Parse(s)
		if err != nil {
			return nil, nil, err
		}
		dnsServers[i] = provider.DNS()
	}

	return dotServers, dnsServers, nil
}

func pickNameAddress(picker picker.DoT, servers []provider.DoTServer,
	ipv6 bool) (name, address string) {
	server := picker.DoTServer(servers)
	ip := picker.DoTIP(server, ipv6)
	address = net.JoinHostPort(ip.String(), fmt.Sprint(server.Port))
	return server.Name, address
}

func onDialError(ctx context.Context, dialErr error,
	dotName, dotAddress string, dialer *net.Dialer,
	picker picker.DNS, ipv6 bool, dnsServers []provider.DNSServer,
	warner log.Warner, metrics metrics.DialMetrics) (
	conn net.Conn, err error) {
	warner.Warn(dialErr.Error())
	metrics.DoTDialInc(dotName, dotAddress, "error")

	if len(dnsServers) == 0 {
		return nil, dialErr
	}

	// fallback on plain DNS if DoT does not work and
	// some plaintext DNS servers are set.
	return dialPlaintext(ctx, dialer, picker, ipv6, dnsServers, warner, metrics)
}

func dialPlaintext(ctx context.Context, dialer *net.Dialer,
	picker picker.DNS, ipv6 bool, dnsServers []provider.DNSServer,
	warner log.Warner, metrics metrics.DialDNSMetrics) (
	conn net.Conn, err error) {
	dnsServer := picker.DNSServer(dnsServers)
	ip := picker.DNSIP(dnsServer, ipv6)

	plainAddr := net.JoinHostPort(ip.String(), "53")

	conn, err = dialer.DialContext(ctx, "udp", plainAddr)
	if err != nil {
		warner.Warn(err.Error())
		metrics.DNSDialInc(plainAddr, "error")
		return nil, err
	}

	metrics.DNSDialInc(plainAddr, "success")
	return conn, nil
}
