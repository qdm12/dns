package localdns

import (
	"net/netip"
	"testing"

	"github.com/miekg/dns"
	"github.com/qdm12/dns/v2/internal/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Parallel()

	settings := Settings{
		Resolvers: []netip.AddrPort{
			netip.AddrPortFrom(netip.MustParseAddr("1.2.3.4"), 53),
		},
		PublicNamesAsLocal: []string{"github.com"},
		Logger:             NewMockLogger(nil),
		TimeoutWarn:        new(true),
	}

	middleware, err := New(settings)
	require.NoError(t, err)

	expectedMiddleware := &Middleware{
		settings: Settings{
			Resolvers:                settings.Resolvers,
			PublicNameserversAsLocal: []netip.Prefix{},
			PublicNamesAsLocal:       settings.PublicNamesAsLocal,
			Logger:                   settings.Logger,
			TimeoutWarn:              settings.TimeoutWarn,
		},
		localChecker: local.New(settings.PublicNamesAsLocal),
	}
	assert.Equal(t, expectedMiddleware, middleware)

	next := dns.HandlerFunc(func(_ dns.ResponseWriter, _ *dns.Msg) {})
	handler := middleware.Wrap(next)

	request := &dns.Msg{Question: []dns.Question{
		{Name: "domain.com."},
	}}
	writer := NewMockResponseWriter(nil)
	handler.ServeDNS(writer, request)

	err = middleware.Stop()
	require.NoError(t, err)
}
