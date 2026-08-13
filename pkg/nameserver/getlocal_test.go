package nameserver

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_addrContainedByAnyPrefix(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		address  netip.Addr
		prefixes []netip.Prefix
		matched  bool
	}{
		"no_prefixes": {
			address: netip.MustParseAddr("8.8.8.8"),
			matched: false,
		},
		"address_not_in_prefixes": {
			address: netip.MustParseAddr("8.8.4.4"),
			prefixes: []netip.Prefix{
				netip.MustParsePrefix("1.1.1.0/24"),
			},
			matched: false,
		},
		"address_in_prefixes": {
			address: netip.MustParseAddr("8.8.8.8"),
			prefixes: []netip.Prefix{
				netip.MustParsePrefix("1.1.1.0/24"),
				netip.MustParsePrefix("8.8.8.0/24"),
			},
			matched: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			matched := addrContainedByAnyPrefix(testCase.address, testCase.prefixes)

			assert.Equal(t, testCase.matched, matched)
		})
	}
}
