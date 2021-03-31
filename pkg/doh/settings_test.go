package doh

import (
	"net/url"
	"testing"
	"time"

	"github.com/qdm12/dns/pkg/cache"
	"github.com/qdm12/dns/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_Settings_setDefaults(t *testing.T) {
	t.Parallel()

	s := Settings{}
	s.setDefaults()

	// Check this otherwise things will blow up if no option is passed.
	assert.GreaterOrEqual(t, len(s.DoHServers), 1)
	assert.GreaterOrEqual(t, len(s.SelfDNS.DoTServers), 1)
	assert.GreaterOrEqual(t, int64(s.Timeout), int64(time.Millisecond))

	expectedSettings := Settings{
		DoHServers: []provider.DoHServer{{
			URL: &url.URL{
				Scheme: "https",
				Host:   "cloudflare-dns.com",
				Path:   "/dns-query",
			},
		}},
		SelfDNS: SelfDNS{
			DoTServers: []provider.DoTServer{
				provider.Cloudflare().DoT(),
			},
		},
		Timeout: 5 * time.Second,
		Port:    53,
		IPv6:    false,
		Cache: cache.Settings{
			Type: cache.NOOP,
		},
	}
	assert.Equal(t, expectedSettings, s)
}

func Test_Settings_SetProviders(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		initialSettings  Settings
		providers        []provider.Provider
		expectedSettings Settings
	}{
		"single provider": {
			initialSettings: Settings{},
			providers: []provider.Provider{
				provider.Google(),
			},
			expectedSettings: Settings{
				DoHServers: []provider.DoHServer{
					provider.Google().DoH(),
				},
				SelfDNS: SelfDNS{
					DoTServers: []provider.DoTServer{
						provider.Google().DoT(),
					},
					DNSServers: []provider.DNSServer{
						provider.Google().DNS(),
					},
				},
			},
		},
		"multiple providers": {
			initialSettings: Settings{},
			providers: []provider.Provider{
				provider.Google(),
				provider.Cloudflare(),
				provider.Quad9(),
			},
			expectedSettings: Settings{
				DoHServers: []provider.DoHServer{
					provider.Cloudflare().DoH(),
					provider.Quad9().DoH(),
					provider.Google().DoH(),
				},
				SelfDNS: SelfDNS{
					DoTServers: []provider.DoTServer{
						provider.Cloudflare().DoT(),
						provider.Quad9().DoT(),
						provider.Google().DoT(),
					},
					DNSServers: []provider.DNSServer{
						provider.Cloudflare().DNS(),
						provider.Quad9().DNS(),
						provider.Google().DNS(),
					},
				},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testCase.initialSettings.SetProviders(
				testCase.providers[0], testCase.providers[1:]...)

			assert.Equal(t, testCase.expectedSettings, testCase.initialSettings)
		})
	}
}
