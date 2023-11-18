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
				Resolver:         ResolverSettings{},
			},
			s: `DoT server settings:
├── Listening address: localhost:53
└── DoT resolver settings:
    ├── DNS over TLS providers:
    ├── Fallback plaintext DNS providers:
    ├── Quey timeout: 0s
    └── Connecting over: IPv4`,
		},
		"non empty settings": {
			settings: ServerSettings{
				ListeningAddress: ptrTo(":8000"),
				Resolver: ResolverSettings{
					DoTProviders: []provider.Provider{
						provider.Cloudflare(),
					},
					DNSProviders: []provider.Provider{
						provider.Cloudflare(), provider.Google(),
					},
					Timeout: time.Second,
					IPv6:    true,
				},
			},
			s: `DoT server settings:
├── Listening address: :8000
└── DoT resolver settings:
    ├── DNS over TLS providers:
    |   └── Cloudflare
    ├── Fallback plaintext DNS providers:
    |   ├── Cloudflare
    |   └── Google
    ├── Quey timeout: 1s
    └── Connecting over: IPv6`,
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := testCase.settings.String()

			assert.Equal(t, testCase.s, s)
		})
	}
}
