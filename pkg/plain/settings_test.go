package plain

import (
	"testing"
	"time"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_Settings_String(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings Settings
		s        string
	}{
		"empty_settings": {
			settings: Settings{
				IPVersion: "ipv4",
			},
			s: `Plain resolver settings:
├── Upstream resolvers:
├── Query timeout: 0s
└── Connecting over: ipv4`,
		},
		"non_empty_settings": {
			settings: Settings{
				UpstreamResolvers: []provider.Provider{
					provider.Cloudflare(),
				},
				Timeout:   time.Second,
				IPVersion: "ipv6",
			},
			s: `Plain resolver settings:
├── Upstream resolvers:
|   └── Cloudflare
├── Query timeout: 1s
└── Connecting over: ipv4 and ipv6`,
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
