package doh

import (
	"net"
	"testing"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func Test_Dialer(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	cloudflare := provider.Cloudflare()

	metrics := NewMockMetrics(ctrl)
	metrics.EXPECT().DoHDialInc(cloudflare.DoH.URL).
		MinTimes(1).MaxTimes(2) // A only or A+AAAA

	dialer, err := New(Settings{
		UpstreamResolvers: []provider.Provider{
			cloudflare,
		},
		Metrics: metrics,
	})
	require.NoError(t, err)

	resolver := &net.Resolver{
		PreferGo: true,
		Dial:     dialer.Dial,
	}

	ips, err := resolver.LookupIPAddr(t.Context(), "github.com")
	require.NoError(t, err)
	require.NotEmpty(t, ips)
}
