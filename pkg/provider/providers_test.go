package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Providers_Get(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		s          string
		provider   Provider
		errWrapped error
		errMessage string
	}{
		"empty string": {
			errWrapped: ErrParseProviderNameUnknown,
			errMessage: "provider does not match any known providers: ",
		},
		"bad provider string": {
			s:          "invalid",
			errWrapped: ErrParseProviderNameUnknown,
			errMessage: "provider does not match any known providers: invalid",
		},
		"cirafamily": {
			s:        "cira family",
			provider: CiraFamily(),
		},
		"ciraprivate": {
			s:        "cira private",
			provider: CiraPrivate(),
		},
		"ciraprotected": {
			s:        "cira protected",
			provider: CiraProtected(),
		},
		"cleanbrowsingadult": {
			s:        "cleanbrowsing adult",
			provider: CleanBrowsingAdult(),
		},
		"cleanbrowsingfamily": {
			s:        "cleanbrowsing family",
			provider: CleanBrowsingFamily(),
		},
		"cleanbrowsingsecurity": {
			s:        "cleanbrowsing security",
			provider: CleanBrowsingSecurity(),
		},
		"cloudflare": {
			s:        "cloudflare",
			provider: Cloudflare(),
		},
		"cloudflarefamily": {
			s:        "cloudflare family",
			provider: CloudflareFamily(),
		},
		"cloudflaremozilla": {
			s:        "cloudflare mozilla",
			provider: CloudflareMozilla(),
		},
		"cloudflaresecurity": {
			s:        "cloudflare security",
			provider: CloudflareSecurity(),
		},
		"google": {
			s:        "google",
			provider: Google(),
		},
		"libredns": {
			s:        "libredns",
			provider: LibreDNS(),
		},
		"quad9": {
			s:        "quad9",
			provider: Quad9(),
		},
		"quad9secured": {
			s:        "quad9 secured",
			provider: Quad9Secured(),
		},
		"quad9unsecured": {
			s:        "quad9 unsecured",
			provider: Quad9Unsecured(),
		},
		"quadrant": {
			s:        "quadrant",
			provider: Quadrant(),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			providers := NewProviders()
			provider, err := providers.Get(testCase.s)

			assert.ErrorIs(t, err, testCase.errWrapped)
			if testCase.errWrapped != nil {
				assert.EqualError(t, err, testCase.errMessage)
			}
			assert.Equal(t, testCase.provider, provider)
		})
	}
}

func Test_Providers_List(t *testing.T) {
	t.Parallel()

	providers := NewProviders()
	listed := providers.List()
	const currentListCount = 17
	assert.GreaterOrEqual(t, len(listed), currentListCount)

	for _, provider := range listed {
		errMessage := "for provider " + provider.Name

		if provider.DoT.Name == "" {
			assert.Empty(t, provider.DoT.IPv4, errMessage)
			assert.Empty(t, provider.DoT.IPv6, errMessage)
		} else {
			assert.NotEmpty(t, append(provider.DoT.IPv4, provider.DoT.IPv6...), errMessage)
			err := checkAddrPorts(provider.DoT.IPv4)
			assert.NoError(t, err, errMessage)
			err = checkAddrPorts(provider.DoT.IPv6)
			assert.NoError(t, err, errMessage)
		}

		if provider.DoH.URL == "" {
			assert.Empty(t, provider.DoH.IPv4, errMessage)
			assert.Empty(t, provider.DoH.IPv6, errMessage)
		} else {
			assert.NotEmpty(t, append(provider.DoH.IPv4, provider.DoH.IPv6...), errMessage)
			err := checkAddresses(provider.DoH.IPv4)
			assert.NoError(t, err, errMessage)
			err = checkAddresses(provider.DoH.IPv6)
			assert.NoError(t, err, errMessage)
		}
	}
}
