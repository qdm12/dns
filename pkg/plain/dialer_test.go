package plain

import (
	"math/rand/v2"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Dialer_pickAddress(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		dialer   Dialer
		dialAddr string
		expected string
	}{
		"address_found": {
			dialer: Dialer{
				addrStrings: map[string]struct{}{
					"1.1.1.1:53": {},
				},
			},
			dialAddr: "1.1.1.1:53",
			expected: "1.1.1.1:53",
		},
		"address_not_found": {
			dialer: Dialer{
				addrStrings: map[string]struct{}{
					"1.1.1.1:53": {},
					"1.0.0.1:53": {},
					"8.8.8.8:53": {},
				},
				addrPorts: []netip.AddrPort{
					netip.AddrPortFrom(netip.MustParseAddr("1.1.1.1"), 53),
					netip.AddrPortFrom(netip.MustParseAddr("1.0.0.1"), 53),
					netip.AddrPortFrom(netip.MustParseAddr("8.8.8.8"), 53),
				},
				randGen: rand.New(rand.NewPCG(1, 1)), //nolint:gosec
			},
			dialAddr: "10.0.0.1:53",
			expected: "8.8.8.8:53",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dialer := testCase.dialer
			address := dialer.pickAddress(testCase.dialAddr)
			assert.Equal(t, testCase.expected, address)
		})
	}
}
