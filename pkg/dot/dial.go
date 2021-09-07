package dot

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/qdm12/dns/pkg/dot/metrics"
	"github.com/qdm12/dns/pkg/log"
	"github.com/qdm12/dns/pkg/provider"
)

type dialFunc func(ctx context.Context, _, _ string) (net.Conn, error)

func newDoTDial(settings ResolverSettings) dialFunc {
	warner := settings.Warner
	metrics := settings.Metrics

	dotServers, dnsServers := settingsToServers(settings)

	dialer := &net.Dialer{
		Timeout: settings.Timeout,
	}

	picker := newPicker()

	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		serverName, serverAddress := pickNameAddress(picker,
			dotServers, settings.IPv6)

		conn, err := dialer.DialContext(ctx, "tcp", serverAddress)
		if err != nil {
			return onDialError(ctx, serverName, serverAddress, dialer,
				picker, settings.IPv6, dnsServers, warner, metrics)
		}

		metrics.DoTDialInc(serverName, serverAddress, "success")

		tlsConf := &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: serverName,
		}
		// TODO handshake? See tls.DialWithDialer
		return tls.Client(conn, tlsConf), nil
	}
}

func settingsToServers(settings ResolverSettings) (
	dotServers []provider.DoTServer,
	dnsServers []provider.DNSServer) {
	dotServers = make([]provider.DoTServer, len(settings.DoTProviders))
	for i := range settings.DoTProviders {
		dotServers[i] = settings.DoTProviders[i].DoT()
	}

	dnsServers = make([]provider.DNSServer, len(settings.DNSProviders))
	for i := range settings.DNSProviders {
		dnsServers[i] = settings.DNSProviders[i].DNS()
	}

	return dotServers, dnsServers
}

func pickNameAddress(picker *picker, servers []provider.DoTServer,
	ipv6 bool) (name, address string) {
	server := picker.DoTServer(servers)
	ip := picker.DoTIP(server, ipv6)
	address = net.JoinHostPort(ip.String(), fmt.Sprint(server.Port))
	return server.Name, address
}

func onDialError(ctx context.Context, dotName, dotAddress string,
	dialer *net.Dialer, picker *picker, ipv6 bool,
	dnsServers []provider.DNSServer, warner log.Warner,
	metrics metrics.DialMetrics) (conn net.Conn, err error) {
	warner.Warn(err.Error())
	metrics.DoTDialInc(dotName, dotAddress, "error")

	if len(dnsServers) == 0 {
		return nil, err
	}

	// fallback on plain DNS if DoT does not work and
	// some plaintext DNS servers are set.
	return dialPlaintext(ctx, dialer, picker, ipv6, dnsServers, warner, metrics)
}

func dialPlaintext(ctx context.Context, dialer *net.Dialer,
	picker *picker, ipv6 bool, dnsServers []provider.DNSServer,
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
