package localdns

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptrTo[T any](v T) *T { return &v }

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
				TimeoutWarn: ptrTo(true),
			},
			s: `Local forwarding middleware settings:
├── Log timeout errors at the warning level: yes
└── Local resolvers:
    ├── 1.2.3.4:53
    └── 9.2.3.4:53`,
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
