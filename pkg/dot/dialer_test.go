package dot

import (
	"math/rand/v2"
	"net"
	"net/netip"
	"slices"
	"strings"
	"testing"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Dialer(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)

	cloudflare := provider.Cloudflare()

	metrics := NewMockMetrics(ctrl)
	possibleAddrs := make([]string, 0, len(cloudflare.DoT.IPv4)+len(cloudflare.DoT.IPv6))
	for _, server := range cloudflare.DoT.IPv4 {
		possibleAddrs = append(possibleAddrs, server.String())
	}
	for _, server := range cloudflare.DoT.IPv6 {
		possibleAddrs = append(possibleAddrs, server.String())
	}
	addrMatcher := &matcherAnyString{
		strings: possibleAddrs,
	}
	metrics.EXPECT().DoTDialInc(cloudflare.DoT.Name, addrMatcher, "success").
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

// matcherAnyString is a [gomock.Matcher] that returns true if
// the argument matches any of the specified strings.
type matcherAnyString struct {
	strings []string
}

func (m matcherAnyString) Matches(x any) bool {
	s, ok := x.(string)
	if !ok {
		return false
	}
	return slices.Contains(m.strings, s)
}

func (m matcherAnyString) String() string {
	return "matches any of: " + strings.Join(m.strings, ", ")
}

func Test_Dialer_pickNameAddress(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		dialer   Dialer
		dialAddr string
		name     string
		address  string
	}{
		"address_found": {
			dialer: Dialer{
				addressToServerName: map[string]string{
					"1.1.1.1:853": "cloudflare-dns.com",
				},
			},
			dialAddr: "1.1.1.1:853",
			name:     "cloudflare-dns.com",
			address:  "1.1.1.1:853",
		},
		"address_not_found": {
			dialer: Dialer{
				addressToServerName: map[string]string{
					"1.1.1.1:853": "cloudflare-dns.com",
					"1.0.0.1:853": "cloudflare-dns.com",
					"8.8.8.8:853": "google-dns.com",
				},
				addrPorts: []netip.AddrPort{
					netip.AddrPortFrom(netip.MustParseAddr("1.1.1.1"), 853),
					netip.AddrPortFrom(netip.MustParseAddr("1.0.0.1"), 853),
					netip.AddrPortFrom(netip.MustParseAddr("8.8.8.8"), 853),
				},
				randGen: rand.New(rand.NewPCG(1, 1)), //nolint:gosec
			},
			dialAddr: "10.0.0.1:853",
			name:     "google-dns.com",
			address:  "8.8.8.8:853",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dialer := testCase.dialer
			name, address := dialer.pickNameAddress(testCase.dialAddr)
			assert.Equal(t, testCase.name, name)
			assert.Equal(t, testCase.address, address)
		})
	}
}
