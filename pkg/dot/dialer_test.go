package dot

import (
	"math/rand/v2"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
