package localdns

import (
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

func runLocalDNS(t *testing.T, handler dns.HandlerFunc,
) (listeningAddress string, runError <-chan error) {
	t.Helper()

	listener, err := net.ListenUDP("udp", nil)
	require.NoError(t, err)
	listeningAddress = listener.LocalAddr().String()

	server := dns.Server{
		PacketConn: listener,
		Handler:    handler,
	}

	readyCh := make(chan struct{})
	server.NotifyStartedFunc = func() {
		close(readyCh)
	}

	runErrorCh := make(chan error)
	go func() {
		runErrorCh <- server.ActivateAndServe()
	}()
	t.Cleanup(func() {
		err := server.Shutdown()
		require.NoError(t, err)
	})

	select {
	case <-readyCh:
	case err := <-runErrorCh:
		t.Fatal("server failed to start:", err)
	}

	return listeningAddress, runErrorCh
}
