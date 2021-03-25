package doh

import (
	"testing"

	"github.com/qdm12/dns/pkg/provider"
	"github.com/stretchr/testify/assert"
)

func Test_Providers(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		initialSettings  settings
		providers        []provider.Provider
		expectedSettings settings
	}{
		"single provider": {
			initialSettings: settings{},
			providers: []provider.Provider{
				provider.Google(),
			},
			expectedSettings: settings{
				providers: []provider.Provider{
					provider.Google(),
				},
				dohServers: []provider.DoHServer{
					provider.Google().DoH(),
				},
			},
		},
		"multiple providers": {
			initialSettings: settings{},
			providers: []provider.Provider{
				provider.Google(),
				provider.Cloudflare(),
				provider.Quad9(),
			},
			expectedSettings: settings{
				providers: []provider.Provider{
					provider.Cloudflare(),
					provider.Quad9(),
					provider.Google(),
				},
				dohServers: []provider.DoHServer{
					provider.Cloudflare().DoH(),
					provider.Quad9().DoH(),
					provider.Google().DoH(),
				},
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			option := Providers(testCase.providers[0], testCase.providers[1:]...)
			option(&testCase.initialSettings)

			assert.Equal(t, testCase.expectedSettings, testCase.initialSettings)
		})
	}
}
