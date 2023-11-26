package localdns

import (
	"net/netip"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_New(t *testing.T) {
	t.Parallel()

	t.Run("resolvers_not_set", func(t *testing.T) {
		t.Parallel()

		settings := Settings{}

		middleware, err := New(settings)
		assert.ErrorIs(t, err, ErrResolversNotSet)
		require.EqualError(t, err, "validating settings: resolvers are not set")
		assert.Nil(t, middleware)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		settings := Settings{
			Resolvers: []netip.AddrPort{
				netip.AddrPortFrom(netip.MustParseAddr("1.2.3.4"), 53),
			},
			Logger: NewMockLogger(nil),
		}

		middleware, err := New(settings)
		require.NoError(t, err)

		expectedMiddleware := &Middleware{
			settings: settings,
		}
		assert.Equal(t, expectedMiddleware, middleware)

		next := dns.HandlerFunc(func(rw dns.ResponseWriter, m *dns.Msg) {})
		handler := middleware.Wrap(next)

		request := &dns.Msg{Question: []dns.Question{
			{Name: "domain.com."},
		}}
		writer := NewMockResponseWriter(nil)
		handler.ServeDNS(writer, request)

		err = middleware.Stop()
		require.NoError(t, err)
	})
}
