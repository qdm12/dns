package localdns

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Settings_String(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings Settings
		s        string
	}{
		"multiple_resolvers": {
			settings: Settings{
				Resolvers: []netip.AddrPort{
					netip.MustParseAddrPort("1.2.3.4:53"),
					netip.MustParseAddrPort("9.2.3.4:53"),
				},
				PublicNameserversAsLocal: []netip.Prefix{
					netip.MustParsePrefix("8.8.8.0/24"),
				},
				TimeoutWarn: new(true),
			},
			s: `Local forwarding middleware settings:
├── Log timeout errors at the warning level: yes
├── Local resolvers:
|   ├── 1.2.3.4:53
|   └── 9.2.3.4:53
└── Public nameserver CIDRs considered local: 8.8.8.0/24`,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := testCase.settings.String()

			assert.Equal(t, testCase.s, s)
		})
	}
}

func Test_Settings_Validate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings Settings
		err      string
	}{
		"valid": {
			settings: Settings{
				PublicNameserversAsLocal: []netip.Prefix{netip.MustParsePrefix("8.8.8.0/24")},
			},
		},
		"invalid_prefix": {
			settings: Settings{
				PublicNameserversAsLocal: []netip.Prefix{{}},
			},
			err: "nameserver public CIDR is not valid",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.settings.Validate()

			if testCase.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.err)
				return
			}

			require.NoError(t, err)
		})
	}
}
