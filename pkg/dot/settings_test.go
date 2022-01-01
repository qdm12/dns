package dot

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ServerSettings_String(t *testing.T) {
	t.Parallel()

	boolPtr := func(b bool) *bool { return &b }

	testCases := map[string]struct {
		settings ServerSettings
		s        string
	}{
		"empty settings": {
			s: `DoT server settings:
├── Listening address: 
└── DoT resolver settings:
    ├── DNS over TLS providers:
    ├── Fallback plaintext DNS providers:
    ├── Quey timeout: 0s
    └── Connecting over: IPv4`,
		},
		"non empty settings": {
			settings: ServerSettings{
				Address: ":8000",
				Resolver: ResolverSettings{
					DoTProviders: []string{
						"cloudflare",
					},
					DNSProviders: []string{
						"cloudflare", "google",
					},
					Timeout: time.Second,
					IPv6:    boolPtr(true),
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
