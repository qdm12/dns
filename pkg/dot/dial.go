package dot

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/qdm12/dns/v2/internal/server"
	"github.com/qdm12/dns/v2/pkg/provider"
)

func newDoTDial(settings ResolverSettings) (dial server.Dial) {
	warner := settings.Warner
	metrics := settings.Metrics

	dotServers, dnsServers := settingsToServers(settings)

	dialer := &net.Dialer{
		Timeout: settings.Timeout,
	}

	picker := settings.Picker

	return func(ctx context.Context, _, _ string) (net.Conn, error) {
		serverName, serverAddress := pickNameAddress(picker,
			dotServers, settings.IPv6)

		conn, err := dialer.DialContext(ctx, "tcp", serverAddress)
		if err != nil {
			return onDialError(ctx, err, serverName, serverAddress, dialer,
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
	dotServers []provider.DoTServer, dnsServers []provider.DNSServer) {
	dotServers = make([]provider.DoTServer, len(settings.DoTProviders))
	for i, provider := range settings.DoTProviders {
		dotServers[i] = provider.DoT
	}

	dnsServers = make([]provider.DNSServer, len(settings.DNSProviders))
	for i, provider := range settings.DNSProviders {
		dnsServers[i] = provider.DNS
	}

	return dotServers, dnsServers
}

func pickNameAddress(picker Picker, servers []provider.DoTServer,
	ipv6 bool) (name, address string) {
	server := picker.DoTServer(servers)
	ip := picker.DoTIP(server, ipv6)
	address = net.JoinHostPort(ip.String(), fmt.Sprint(server.Port))
	return server.Name, address
}

func onDialError(ctx context.Context, dialErr error,
	dotName, dotAddress string, dialer *net.Dialer,
	picker Picker, ipv6 bool, dnsServers []provider.DNSServer,
	warner Warner, metrics Metrics) (
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
	picker Picker, ipv6 bool, dnsServers []provider.DNSServer,
	warner Warner, metrics Metrics) (
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
