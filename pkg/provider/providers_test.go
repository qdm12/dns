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
		testCase := testCase
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
	assert.Len(t, listed, 16)

	for _, provider := range listed {
		errMessage := "for provider " + provider.DoT.Name

		assert.NotNil(t, provider.DoT.IPv6, errMessage)
		assert.NotEmpty(t, provider.DoT.Name, errMessage)
		err := checkAddrPorts(provider.DoT.IPv4)
		assert.NoError(t, err, errMessage)
		err = checkAddrPorts(provider.DoT.IPv6)
		assert.NoError(t, err, errMessage)

		assert.NotNil(t, provider.DoH, errMessage)
	}
}
