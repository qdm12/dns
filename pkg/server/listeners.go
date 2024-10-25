package server

import (
	"context"
	"fmt"
	"net"
	"strings"
)

// setupListeners finds a TCP and a UDP listener for the given
// listening address. In particular, it will find a randomly available
// port for both TCP and UDP if the listening address is with port 0.
func setupListeners(ctx context.Context, listeningAddress string) (
	udpListener *net.UDPConn, tcpListener *net.TCPListener, err error,
) {
	tcpListeningAddress, err := net.ResolveTCPAddr("tcp", listeningAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("resolving tcp listening address: %w", err)
	}
	udpListeningAddress, err := net.ResolveUDPAddr("udp", listeningAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("resolving udp listening address: %w", err)
	}
	if tcpListeningAddress.String() != udpListeningAddress.String() {
		panic("tcp and udp listening addresses are different")
	}

	for ctx.Err() == nil {
		tcpListener, err = net.ListenTCP("tcp", tcpListeningAddress)
		if err != nil {
			return nil, nil, fmt.Errorf("creating TCP listener: %w", err)
		}

		udpListener, err = net.ListenUDP("udp", udpListeningAddress)
		if err != nil {
			_ = tcpListener.Close()
			if tcpListeningAddress.Port == 0 &&
				strings.HasSuffix(err.Error(), "address already in use") {
				continue // try finding another port
			}
			return nil, nil, fmt.Errorf("creating UDP listener: %w", err)
		}

		return udpListener, tcpListener, nil
	}

	return nil, nil, ctx.Err()
}
