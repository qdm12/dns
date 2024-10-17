package dot

import (
	"testing"
	"time"

	"github.com/qdm12/dns/v2/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_ServerSettings_String(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		settings ServerSettings
		s        string
	}{
		"empty settings": {
			settings: ServerSettings{
				ListeningAddress: ptrTo("localhost:53"),
				Resolver: ResolverSettings{
					IPVersion: "ipv4",
				},
			},
			s: `DoT server settings:
├── Listening address: localhost:53
└── DoT resolver settings:
    ├── Upstream resolvers:
    ├── Query timeout: 0s
    └── Connecting over: ipv4`,
		},
		"non empty settings": {
			settings: ServerSettings{
				ListeningAddress: ptrTo(":8000"),
				Resolver: ResolverSettings{
					UpstreamResolvers: []provider.Provider{
						provider.Cloudflare(),
					},
					Timeout:   time.Second,
					IPVersion: "ipv6",
				},
			},
			s: `DoT server settings:
├── Listening address: :8000
└── DoT resolver settings:
    ├── Upstream resolvers:
    |   └── Cloudflare
    ├── Query timeout: 1s
    └── Connecting over: ipv6`,
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
